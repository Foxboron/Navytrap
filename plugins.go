package navytrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
)

type cmd func(string) error
type Cmd func(*Connection, *Parsed)

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

var (
	Privmsgs = make(map[string]Cmd)
	Joins    = make(map[string]Cmd)
	// Part    map[string]Cmd
	// Kick    map[string]Cmd
	// Mode    map[string]Cmd
	Signals = make(map[string]Cmd)
	Events  = make(map[string]Cmd)
)

func RegisterPrivmsg(s string, f Cmd) {
	fmt.Println(Privmsgs)
	Privmsgs[s] = f
}

func RegisterEvent(event string, f Cmd) {
	fmt.Println(Events)
	Events[event] = f
}

func Listen() {
	// PRIVMSG
	go func() {
		// TODO: Find a better way to handle this
		for {
			select {
			case p := <-Signal:
				if event, ok := Events[p.Msg.Cmd]; ok {
					go event(p.Conn, p.Msg)
				}
				if p.Msg.Cmd == "PRIVMSG" {
					for k, v := range Privmsgs {
						matched, _ := regexp.MatchString("^"+k, p.Msg.Args[1])
						if matched {
							go v(p.Conn, p.Msg)
						}
					}
				}
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
	runFunc, ok := run.(func() error)
	if !ok {
		return fmt.Errorf("found Run but type is %T instead of func() error", run)
	}
	go runFunc()
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
	Listen()
	return nil
}

var Factoids = make(map[string]string)

func init() {
	initPlugins()
	// Join channels
	fmt.Println("lol")

	RegisterEvent("266", func(net *Connection, p *Parsed) {
		fmt.Println("GOT 266 EVENT!")
		for _, c := range net.Channels {
			fmt.Println(c)
			net.JoinChannel(c)
		}
	})

	RegisterPrivmsg("!\\w* is .*", func(n *Connection, p *Parsed) {
		addition := strings.SplitN(p.Args[1], " is ", 2)
		Factoids[addition[0][1:]] = addition[1]
	})

	RegisterPrivmsg("!.*", func(n *Connection, p *Parsed) {
		is_fact := strings.Split(p.Args[1], " ")[0][1:]

		if fact, ok := Factoids[is_fact]; ok {
			n.WriteChannel(p.Channel, fact)
		}
	})

	RegisterPrivmsg("!give \\w* .*", func(n *Connection, p *Parsed) {
		give := strings.SplitN(p.Args[1], " ", 3)

		if fact, ok := Factoids[give[2]]; ok {
			n.WriteChannel(p.Channel, give[1]+": "+fact)
		}
	})
}
