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
	"log"
	"net"
	"os"
	"sync"
	"time"
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
	Servers = make(map[string]Connection)
)

var serverChan = make(chan *Parsed) // output from server
var done = make(chan struct{})

var Conn *Connection

func mustWriteln(w io.Writer, s string) {
	if _, err := fmt.Fprint(w, s+"\r\n"); err != nil {
		log.Fatal(err)
	}
}

func mustWritef(w io.Writer, form string, args ...interface{}) {
	mustWriteln(w, fmt.Sprintf(form, args...))
}

// listenServer scans for server messages on conn and sends
// them on serverChan.
func listenServer(conn *Connection) {
	in := bufio.NewScanner(conn)
	for in.Scan() {
		if p, err := Parse(in.Text()); err != nil {
			log.Print("parse error:", err)
		} else {
			serverChan <- p
		}
	}
	close(done)
}

func connServer(server string, port string, useTLS bool) net.Conn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		log.Print(err)
	}
	if useTLS {
		return tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
	}
	return conn
}

func handleServer(conn *Connection, p *Parsed) {
	switch p.Cmd {
	case "PING":
		mustWritef(conn, "PONG %s", p.Args[0])
	case "PONG":
		break
	case "PART":
		fallthrough
	default:
		sendPkg := &Pkg{Conn: conn, Msg: p}
		Signal <- sendPkg
	}
}

func login(conn *Connection, server string, pass string, name string) {
	if pass != "" {
		mustWritef(conn, "PASS %s", pass)
	}
	mustWritef(conn, "NICK %s", name)
	mustWritef(conn, "USER %s localhost %s :%s", name, name, name)
}

func run(conn *Connection, server string, channels []string) {
	go listenServer(conn)
	ticker := time.NewTicker(1 * time.Minute)
	for {
	loop:
		select {
		case <-done:
			break loop
		case <-ticker.C: // FIXME: ping timeout check
			mustWritef(conn, "PING %s", server)
		case s := <-serverChan:
			// if s.Cmd == "266" {
			// 	for _, c := range channels {
			// 		joinChannel(conn, c)
			// 	}
			// }
			handleServer(conn, s)
		}
	}
}

func joinChannel(conn *Connection, c string) {
	mustWritef(conn, "JOIN %s", c)
}

func CreateConnections() {
	for _, server := range config.Servers {
		conn := connServer(server.Address, server.Port, server.Tls)
		Conn = &Connection{Conn: conn, Channels: server.Channels}
		defer Conn.Close()
		login(Conn, server.Address, os.Getenv(""), config.Nick)
		go run(Conn, server.Address, server.Channels)
		wg.Add(1)
	}
	wg.Wait()
}
