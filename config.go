package navytrap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Server struct {
	Address  string   `json:"address"`
	Port     string   `json:"port"`
	Tls      bool     `json:"tls"`
	Channels []string `json:"channels"`
}

type Config struct {
	Nick     string   `json:"nick"`
	Realname string   `json:"realname"`
	Servers  []Server `json:"servers"`
}

var config Config
var configDir = "./config.json"

func parseFlags(params string) {
	// nick := flag.String("n", "", "IRC nick ($USER)")
	// pass := flag.String("k", "", "Read password from variable (e.g. IIPASS)")
	// port := flag.String("p", "", "Server port (6667/TLS: 6697)")
	// realName := flag.String("f", "", "Real name (nick)")
	// server := flag.String("s", "", "Server to connect to")
	// tls := flag.Bool("t", false, "Use TLS")
	// channel := flag.String("c", "", "Default Channel")
	// flag.Parse()

	// initialize set of channel-less commands
	// for _, s := range glMsgs {
	// 	globalCmds[s] = struct{}{}
	// }

	// if *port == "" {
	// 	if *tls {
	// 		*port = "6697"
	// 	} else {
	// 		*port = "6667"
	// 	}
	// }

	// usr, err := user.Current()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if *nick == "" {
	// 	*nick = usr.Username
	// }
	// if *realName == "" {
	// 	*realName = *nick
	// }
	// clientNick = *nick
}

func ParseConfig(configFile string) Config {
	file, e := ioutil.ReadFile(configFile)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	var config Config
	json.Unmarshal(file, &config)
	return config
}

func ParseModuleConfig(i interface{}) interface{} {
	file, e := ioutil.ReadFile(configDir)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(file, i)
	return i

}

func init() {
	config = ParseConfig(configDir)
}
