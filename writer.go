package navytrap

import (
	"fmt"
	"log"
	"net"
)

type ConnectionInterface interface {
}

type Connection struct {
	net.Conn
	Channels []string
	Server   string
	Tls      bool
	Port     string
}

func (c *Connection) Writeln(s string) {
	fmt.Println(s)
	if _, err := fmt.Fprint(c, s+"\r\n"); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) Writef(form string, args ...interface{}) {
	c.Writeln(fmt.Sprintf(form, args...))
}

func (c *Connection) WriteChannel(ch string, s string) {
	c.Writef("PRIVMSG %s :%s", ch, s)
}

func (c *Connection) JoinChannel(ch string) {
	c.Writef("JOIN %s", ch)
}
