package loader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"regexp"

	net "github.com/Foxboron/Navytrap/net"
	"github.com/Foxboron/Navytrap/parser"
)

type cmd func(string) error

var cmdsMap = make(map[string]cmd)

// Horrible hack :c

type Cmd func(*net.Connection, parser.Parsed)

type Cmds struct {
	Privmsg map[string]Cmd
}

var Cc = &Cmds{Privmsg: make(map[string]Cmd)}

func (c *Cmds) RegisterPrivmsg(s string, f Cmd) {
	c.Privmsg[s] = f
}

func (c *Cmds) Listen() {
	// PRIVMSG
	go func() {
		// TODO: Find a better way to handle this
		for {
			select {
			case p := <-net.Privmsg:
				parsed := p.Parsed
				conn := p.Conn
				for k, v := range c.Privmsg {
					matched, _ := regexp.MatchString("^"+k, parsed.Args[1])
					if matched {
						go v(conn, parsed)
					}
				}
			}
		}
	}()
}

func LoadPlugins(name string) error {
	fmt.Println(name)
	p, err := plugin.Open(name)
	if err != nil {
		return fmt.Errorf("could not open plugin: %v", err)
	}
	run, err := p.Lookup("Run")
	if err != nil {
		return fmt.Errorf("could not find Run function: %v", err)
	}
	runFunc, ok := run.(func(*Cmds) error)
	if !ok {
		return fmt.Errorf("found Run but type is %T instead of func() error", run)
	}
	go runFunc(Cc)
	return nil
}

func RunPlugins() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not find current directory: %v", err)
	}

	pluginsDir := filepath.Join(wd, "plugins")

	dir, err := os.Open(pluginsDir)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		if filepath.Ext(name) == ".so" {
			err := LoadPlugins(pluginsDir + "/" + name)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	Cc.Listen()
	return nil
}
