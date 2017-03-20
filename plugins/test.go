package main

import (
	"fmt"

	"github.com/Foxboron/Navytrap/loader"
	"github.com/Foxboron/Navytrap/net"
	"github.com/Foxboron/Navytrap/parser"
)

func Run(c *loader.Cmds) error {
	c.RegisterPrivmsg("!test", func(n *net.Connection, p parser.Parsed) {
		fmt.Println("wehey")
		n.WriteChannel(p.Channel, "Another test!")
	})
	return nil
}
