/**
This package is essentially a copypasta from the iii project
with the patchset from Wolfgang Corcoran-Mathe. Fitting this into
Navytrap is an on-going project

Problems:
We don't really know when the connection dies. ¯\_(ツ)_/¯
**/

package navytrap

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// TODO: Since we use channels, and
// the plugin system is not aware of the
// connection. We pass around Pkg as a form
// of contexts. This should be redone...
type Pkg struct {
	Conn *Connection
	Msg  *Parsed
}

var (
	wg      sync.WaitGroup
	Servers = make(map[string]*Connection)
)

// var serverChan = make(chan *Parsed) // output from server
var done = make(chan struct{})

func mustWriteln(w io.Writer, s string) {
	if _, err := fmt.Fprint(w, s+"\r\n"); err != nil {
		logger.Fatal(err)
	}
}

func mustWritef(w io.Writer, form string, args ...interface{}) {
	mustWriteln(w, fmt.Sprintf(form, args...))
}

// listenServer scans for server messages on conn and sends
// them on serverChan.
func (c *Connection) listenServer() {
	in := bufio.NewScanner(c.Conn)
	for in.Scan() {
		if p, err := Parse(in.Text()); err != nil {
			logger.Print("parse error:", err)
		} else {
			c.ServerChan <- p
		}
	}
	close(done)
}

func connServer(server string, port string, useTLS bool) net.Conn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server+":"+port)
	if err != nil {
		logger.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logger.Fatal(err)
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		logger.Print(err)
	}
	if useTLS {
		return tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
	}
	return conn
}

func (c *Connection) handleServer(p *Parsed) {
	switch p.Cmd {
	case "PING":
		mustWritef(c.Conn, "PONG %s", p.Args[0])
	case "PONG":
		break
	case "PART":
		fallthrough
	default:
		sendPkg := &Pkg{Conn: c, Msg: p}
		Signal <- sendPkg
	}
}

func (c *Connection) login(server string, pass string, name string) {
	if pass != "" {
		mustWritef(c.Conn, "PASS %s", pass)
	}
	mustWritef(c.Conn, "NICK %s", name)
	mustWritef(c.Conn, "USER %s localhost %s :%s", name, name, name)
}

func (c *Connection) run(server string, channels []string) {
	go c.listenServer()
	ticker := time.NewTicker(1 * time.Minute)
	for {
	loop:
		select {
		case <-done:
			break loop
		case <-ticker.C: // FIXME: ping timeout check
			c.Writef("PING %s", server)
		case s := <-c.ServerChan:
			c.handleServer(s)
		}
	}
}

func joinChannel(conn *Connection, c string) {
	mustWritef(conn.Conn, "JOIN %s", c)
	logger.WithFields(log.Fields{
		"channel": c,
	}).Info("Joined channel")
}

func CreateConnections() {
	for _, server := range config.Servers {
		conn := connServer(server.Address, server.Port, server.Tls)
		Conn := &Connection{Conn: conn, Channels: server.Channels, ServerChan: make(chan *Parsed)}
		defer Conn.Close()
		Conn.login(server.Address, os.Getenv(""), config.Nick)
		go Conn.run(server.Address, server.Channels)
		Servers[server.Address] = Conn
		wg.Add(1)
	}
	wg.Wait()
}
