package navytrap

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"regexp"

	log "github.com/Sirupsen/logrus"
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
	logger.WithFields(log.Fields{
		"event": s,
		"type":  "privmsg",
	}).Debug("Registered event")
	Privmsgs[s] = f
}

func RegisterEvent(event string, f Cmd) {
	logger.WithFields(log.Fields{
		"event": event,
		"type":  "privmsg",
	}).Debug("Registered event")
	Events[event] = f
}

func Listen() {
	// PRIVMSG
	go func() {
		// TODO: Find a better way to handle this
		for {
			select {
			case p := <-Signal:
				logger.WithFields(log.Fields{
					"signal": p.Msg.Cmd,
				}).Debug("Handled signal")
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
	logger.WithFields(log.Fields{
		"plugin": name}).Debug("Loading module...")

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
		logger.Error(err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		logger.Error(err)
	}

	for _, name := range names {
		if filepath.Ext(name) == ".so" {
			err := LoadPlugin(pluginsDir + "/" + name)
			if err != nil {
				logger.Error(err)
			}
		}
	}
	Listen()
	return nil
}

var Factoids = make(map[string]string)

func init() {
	initPlugins()
}
