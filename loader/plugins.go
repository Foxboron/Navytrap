package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"regexp"

	"github.com/Foxboron/Navytrap/parser"
)

type cmd func(string) error

var cmdsMap = make(map[string]cmd)

type Cmd func(parser.Parsed)

type Cmds struct {
	Cmds map[string]Cmd
}

var Cc = &Cmds{Cmds: make(map[string]Cmd)}

func (c *Cmds) RegisterCmd(s string, f Cmd) {
	c.Cmds[s] = f
}

func (c *Cmds) FindCmd(s string) Cmd {
	for k, v := range c.Cmds {
		matched, _ := regexp.MatchString("^"+k, s)
		if matched {
			return v
		}
	}
	return nil
}

func RunPlugins() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not find current directory: %v", err)
	}
	pluginsDir := filepath.Join(wd, "plugins")
	p, err := plugin.Open(pluginsDir + "/plugin.so")
	if err != nil {
		fmt.Println(fmt.Errorf("could not open plugin: %v", err))

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
	if err := runFunc(Cc); err != nil {
		return fmt.Errorf("plugin failed with error %v", err)
	}
	return nil
}
