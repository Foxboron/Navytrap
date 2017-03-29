package navytrap

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Factoid struct {
	gorm.Model
	Key   string
	Value string
}

type Quote struct {
	gorm.Model
	Nick  string
	Quote string
}

type AuthConfig struct {
	Apiauth map[string]string `json:"apiauth"`
}

// Hash is Server+channel+nick
// Because i'm lazy
var Grabs = make(map[string]string)

// Default plugins
func init() {

	db, err := gorm.Open("sqlite3", "navytrap.db")
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Factoid{})
	db.AutoMigrate(&Quote{})

	RegisterEvent("266", func(net *Connection, p *Parsed) {
		for _, c := range net.Channels {
			net.JoinChannel(c)
		}
	})

	RegisterPrivmsg("!\\w* is .*", func(n *Connection, p *Parsed) {
		addition := strings.SplitN(p.Args[1], " is ", 2)
		db.Create(&Factoid{Key: addition[0][1:], Value: addition[1]})
		n.WriteChannel(p.Channel, p.Nick+": Done!")
	})

	RegisterPrivmsg("!\\w*$", func(n *Connection, p *Parsed) {
		var fact Factoid

		is_fact := strings.Split(p.Args[1], " ")[0][1:]

		db.First(&fact, "key = ?", is_fact)
		if fact.Value != "" {
			n.WriteChannel(p.Channel, fact.Value)
		}
	})

	RegisterPrivmsg("!give \\w* .*", func(n *Connection, p *Parsed) {
		var fact Factoid
		give := strings.SplitN(p.Args[1], " ", 3)

		db.First(&fact, "key = ?", give[2])
		if fact.Value != "" {
			n.WriteChannel(p.Channel, give[1]+": "+fact.Value)
		}
	})

	RegisterPrivmsg("*", func(n *Connection, p *Parsed) {
		Grabs[n.Server+p.Channel+p.Nick] = p.Args[1]
	})

	RegisterPrivmsg("!grab \\w*", func(n *Connection, p *Parsed) {
		nick := strings.SplitN(p.Args[1], " ", 2)

		if last_msg, ok := Grabs[n.Server+p.Channel+nick[1]]; ok {
			db.Create(&Quote{Nick: nick[1], Quote: last_msg})
			n.WriteChannel(p.Channel, p.Nick+": Tada!")
		}
	})

	RegisterPrivmsg("!rq \\w*", func(n *Connection, p *Parsed) {
		var quote Quote
		nick := strings.SplitN(p.Args[1], " ", 2)
		db.Order("RANDOM()").First(&quote, "nick = ?", strings.TrimSpace(nick[1]))
		if quote.Quote != "" {
			n.WriteChannel(p.Channel, "<"+quote.Nick+"> "+quote.Quote)
		}
	})

	RegisterPrivmsg("https?://", func(n *Connection, p *Parsed) {
		// resp, err := http.Get(p.Msg)
		resp, err := http.Get(p.Msg)
		if err != nil {
			logger.Error("Could not read http request")
		}
		body, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(body)
		re := regexp.MustCompile("<title>(.*)</title>")
		title := re.FindStringSubmatch(bodyString)
		fmt.Println(title)
		if len(title) >= 2 {
			n.WriteChannel(p.Channel, title[1])
		}

	})

	// HTTP API
	var authconfig AuthConfig
	ParseModuleConfig(&authconfig)

	if authconfig.Apiauth != nil {

		r := gin.New()
		r.Use(ginrus.Ginrus(logger, time.RFC3339, true))

		auth := r.Group("/", gin.BasicAuth(authconfig.Apiauth))

		auth.GET("/:server", func(c *gin.Context) {
			server := c.Param("server")

			if s, ok := Servers[server]; ok {
				chans := s.Channels
				c.JSON(200, gin.H{"channels": chans})
			}
		})

		auth.POST("/privmsg/:server", func(c *gin.Context) {
			msg := c.PostForm("msg")
			channel := c.PostForm("channel")

			server := c.Param("server")
			if s, ok := Servers[server]; ok {
				s.WriteChannel(channel, msg)
			}
		})

		auth.POST("/kick/:server/:nick", func(c *gin.Context) {
			msg := c.PostForm("msg")
			channel := c.Param("channel")
			nick := c.Param("nick")
			server := c.Param("server")
			if s, ok := Servers[server]; ok {
				s.Writef("KICK %s %s: %s", channel, nick, msg)
			}
		})

		go r.Run(":8080")
	}
}
