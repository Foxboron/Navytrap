package net

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	config "github.com/Foxboron/Navytrap/config"
	"github.com/Foxboron/Navytrap/loader"
	"github.com/Foxboron/Navytrap/parser"
)

var msgChan = make(chan string)           //dunno yet
var serverChan = make(chan parser.Parsed) // output from server
var done = make(chan struct{})

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
func listenServer(conn net.Conn) {
	in := bufio.NewScanner(conn)
	for in.Scan() {
		if p, err := parser.Parse(in.Text()); err != nil {
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

func handleServer(conn net.Conn, p parser.Parsed) {
	switch p.Cmd {
	case "PING":
		mustWritef(conn, "PONG %s", p.Args[0])
	case "PONG":
		break
	case "PART":
		fallthrough
	default:
		if p.Cmd == "PRIVMSG" {
			r := loader.Cc.FindCmd(p.Args[1])
			fmt.Println(r)
			if r != nil {
				p.Conn = conn
				r(p)
			}
		}
	}
}

func login(conn net.Conn, server string, pass string, name string) {
	if pass != "" {
		mustWritef(conn, "PASS %s", pass)
	}
	mustWritef(conn, "NICK %s", name)
	mustWritef(conn, "USER %s localhost %s :%s", name, name, name)
}

func run(conn net.Conn, server string, channels []string) {
	go listenServer(conn)
	ticker := time.NewTicker(1 * time.Minute)
	for {
	loop:
		select {
		case <-done:
			for range msgChan {
				// TODO: IDK fam
				// drain remaining
			}
			break loop
		case <-ticker.C: // FIXME: ping timeout check
			mustWritef(conn, "PING %s", server)
		case s := <-serverChan:
			if s.Cmd == "266" {
				for _, c := range channels {
					joinChannel(conn, c)
				}
			}
			handleServer(conn, s)
		}
	}
}

func joinChannel(conn net.Conn, c string) {
	mustWritef(conn, "JOIN %s", c)
}

func Connection(config config.Config) {
	conn := connServer(config.Servers.Address, config.Servers.Port, config.Servers.Tls)
	defer conn.Close()
	login(conn, config.Servers.Address, os.Getenv(""), config.Nick)
	run(conn, config.Servers.Address, config.Servers.Channels)
}
