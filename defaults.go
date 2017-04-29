package navytrap

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

type Data struct {
	Nick string
	User string
}

// Hash is Server+channel+nick
// Because i'm lazy
var Grabs = make(map[string]string)

// We need a lock, as N threads could be writing to this map
var lock = sync.RWMutex{}

// Factoid parser
func parse(s string, d *Data) string {
	// First pass to sort out
	// data such as .User and .Nick
	re := regexp.MustCompile("(\\.\\w+)")
	s = re.ReplaceAllString(s, "{{$1}}")

	tmpl, _ := template.New("firstPass").Parse(s)

	var result bytes.Buffer
	tmpl.Execute(&result, d)
	firstPass := result.String()

	// Format template
	// Check for random part
	// (something|something)
	// -> {{Random "something" "something"}}
	if matched, _ := regexp.MatchString("\\(.*\\)", firstPass); matched {
		re := regexp.MustCompile("\\([\\w ]*\\|+[\\w \\|]+\\)")
		parsing := re.FindAllString(firstPass, -1)
		for _, v := range parsing {
			var newS string
			old := v
			v = strings.Replace(v, "(", "{{Random \"", 1)
			v = strings.Replace(v, "|", "\" \"", -1)
			v = strings.Replace(v, ")", "\"}}", 1)
			newS += v
			firstPass = strings.Replace(firstPass, old, newS, 1)
		}
	}

	funcMap := template.FuncMap{
		"Random": func(s ...string) string {
			rand.Seed(time.Now().Unix())
			return s[rand.Intn(len(s))]
		},
	}

	// Second pass for the random nick or other functions
	tmpl, _ = template.New("secondPass").Funcs(funcMap).Parse(string(firstPass))
	var second bytes.Buffer
	tmpl.Execute(&second, d)
	finish := second.String()

	return finish
}

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
			data := &Data{User: p.Nick, Nick: config.Nick}
			ret := parse(fact.Value, data)

			if strings.HasPrefix(ret, "<action>") {
				n.WriteChannel(p.Channel, "\001ACTION "+strings.Replace(ret, "<action>", "", 1)+"\001")
			} else if strings.HasPrefix(ret, "<reply>") {
				n.WriteChannel(p.Channel, strings.Replace(ret, "<reply>", "", 1))
			} else {
				n.WriteChannel(p.Channel, ret)
			}
		}
	})

	RegisterPrivmsg("!give \\w* .*", func(n *Connection, p *Parsed) {
		var fact Factoid
		give := strings.SplitN(p.Args[1], " ", 3)

		db.First(&fact, "key = ?", give[2])
		if fact.Value != "" {
			data := &Data{User: give[1], Nick: config.Nick}
			ret := parse(fact.Value, data)

			if strings.HasPrefix(ret, "<action>") {
				n.WriteChannel(p.Channel, "\001ACTION "+strings.Replace(ret, "<action>", "", 1)+"\001")
			} else if strings.HasPrefix(ret, "<reply>") {
				n.WriteChannel(p.Channel, strings.Replace(ret, "<reply>", "", 1))
			} else {
				n.WriteChannel(p.Channel, give[1]+": "+ret)
			}
		}
	})

	RegisterPrivmsg("*", func(n *Connection, p *Parsed) {
		lock.Lock()
		defer lock.Unlock()
		Grabs[n.Server+p.Channel+p.Nick] = p.Args[1]
	})

	RegisterPrivmsg("!grab \\w*", func(n *Connection, p *Parsed) {
		nick := strings.SplitN(p.Args[1], " ", 2)

		lock.RLock()
		defer lock.RUnlock()
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

	RegisterPrivmsg("!multirq \\w*", func(n *Connection, p *Parsed) {
		var quotes string
		nick := strings.SplitN(p.Msg, " ", 2)
		for i := 1; i <= 5; i++ {
			var quote Quote
			db.Order("RANDOM()").First(&quote, "nick = ?", strings.TrimSpace(nick[1]))
			if quote.Quote != "" {
				quotes += " <" + quote.Nick + "> " + quote.Quote
			}
		}
		n.WriteChannel(p.Channel, quotes)
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
