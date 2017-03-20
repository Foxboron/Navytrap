package navytrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
)

type cmd func(string) error

var cmdsMap = make(map[string]cmd)

var (
	// All signals goes through this thread
	Signal = make(chan *Pkg)

	// Events we can subscribe to
	Privmsg = make(chan *Pkg)
	Join    = make(chan *Pkg)
	Part    = make(chan *Pkg)
	Kick    = make(chan *Pkg)
	Mode    = make(chan *Pkg)

	// Misc events we will be looking at
	// as an example;
	// event: 266 - join channels
	Misc = make(chan *Pkg)
)

// Horrible hack :c

type Cmd func(*Connection, Parsed)

type Cmds struct {
	Privmsg map[string]Cmd
	Join    map[string]Cmd
	Part    map[string]Cmd
	Kick    map[string]Cmd
	Mode    map[string]Cmd
	Event   map[string]Cmd
	Signal  map[string]Cmd
}

var Cc = &Cmds{Privmsg: make(map[string]Cmd)}

func (c *Cmds) RegisterPrivmsg(s string, f Cmd) {
	c.Privmsg[s] = f
}

func (c *Cmds) RegisterEvent(event string, f Cmd) {
	if c.Event == nil {
		c.Event = make(map[string]Cmd)
	}
	c.Event[event] = f
}

func (c *Cmds) Listen() {
	// PRIVMSG
	go func() {
		// TODO: Find a better way to handle this
		for {
			select {
			case p := <-Signal:
				fmt.Println(p.Conn)
				fmt.Println(p.Msg.Cmd)
				if event, ok := c.Event[p.Msg.Cmd]; ok {
					go event(p.Conn, p.Msg)
				}
				// parsed := p.Parsed
				// conn := p.Conn
				// for k, v := range c.Privmsg {
				// 	matched, _ := regexp.MatchString("^"+k, parsed.Args[1])
				// 	if matched {
				// 		go v(conn, parsed)
				// 	}
				// }
			}
		}
	}()
}

func LoadPlugin(name string) error {
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

func initPlugins() error {
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
			err := LoadPlugin(pluginsDir + "/" + name)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	Cc.Listen()
	return nil
}

func init() {
	initPlugins()
	// Join channels
	Cc.RegisterEvent("266", func(net *Connection, p Parsed) {
		fmt.Println("GOT 266 EVENT!")
		for _, c := range net.Channels {
			fmt.Println(c)
			net.JoinChannel(c)
		}
	})
}
