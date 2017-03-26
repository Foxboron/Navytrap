package navytrap

import (
	"fmt"

	"github.com/foxboron/navytrap"
)

func Run() error {
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
	return nil
}
