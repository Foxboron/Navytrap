package main

import (
	"fmt"

	"github.com/foxboron/navytrap"
	_ "github.com/foxboron/navytrap/navytrap"
)

func main() {
	navytrap.RegisterPrivmsg("!test", func(n *navytrap.Connection, p *navytrap.Parsed) {
		fmt.Println("wehey")
		n.WriteChannel(p.Channel, "Another test!")
	})

	// navytrap.RegisterPrivmsg("!\w* is .*", func(n *Connection, p *Parsed) {
	// 	fmt.Println(p.Msg)
	// })

	// navytrap.RegisterPrivmsg("!.*", func(n *Connection, p *Parsed) {
	// 	fmt.Println(p.Msg)
	// })
}
