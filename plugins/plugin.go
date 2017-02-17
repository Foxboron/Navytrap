package main

import loader "github.com/Foxboron/Navytrap/loader"
import parser "github.com/Foxboron/Navytrap/parser"
import net "github.com/Foxboron/Navytrap/net"

func Run(c *loader.Cmds) error {
	c.RegisterCmd("!test", func(p parser.Parsed) {
		net.WriteChannel(p.Conn, p.Channel, "test")
	})
	return nil
}
