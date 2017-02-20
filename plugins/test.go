package main

import (
	"fmt"

	loader "github.com/Foxboron/Navytrap/loader"
	parser "github.com/Foxboron/Navytrap/parser"

	net "github.com/Foxboron/Navytrap/net"
)

func Run(c *loader.Cmds) error {
	c.RegisterPrivmsg("!anothertest", func(n *net.Connection, p parser.Parsed) {
		fmt.Println("wehey")
		n.WriteChannel(p.Channel, "Another test!")
	})
	return nil
}
