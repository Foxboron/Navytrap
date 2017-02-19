package main

import loader "github.com/Foxboron/Navytrap/loader"
import parser "github.com/Foxboron/Navytrap/parser"
import net "github.com/Foxboron/Navytrap/net"

func Run(c *loader.Cmds) error {
	c.RegisterPrivmsg("!anothertest", func(n *net.Connection, p parser.Parsed) {
		n.WriteChannel(p.Channel, "Another test!")
	})
	return nil
}
