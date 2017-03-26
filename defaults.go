package navytrap

import (
	"strings"

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

}
