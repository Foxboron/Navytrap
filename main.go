package navytrap

import "fmt"

// commands to be logged to server output
var glMsgs = [...]string{"ERROR", "NICK", "QUIT"}

var clientNick string
var clientRealName string

func main() {
	fmt.Println("test")
	CreateConnections()
}

func navytrap() {
	fmt.Println("test")
	CreateConnections()
}
