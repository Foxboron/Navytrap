package main

import (
	config "github.com/Foxboron/Navytrap/config"
	loader "github.com/Foxboron/Navytrap/loader"
	net "github.com/Foxboron/Navytrap/net"
)

// commands to be logged to server output
var glMsgs = [...]string{"ERROR", "NICK", "QUIT"}

var clientNick string
var clientRealName string

func main() {
	loader.RunPlugins()
	config := config.ParseConfig("./config.json")
	net.Connection(config)
}
