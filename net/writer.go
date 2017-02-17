package net

import (
	"fmt"
	"io"
	"log"
)

func Writeln(w io.Writer, s string) {
	if _, err := fmt.Fprint(w, s+"\r\n"); err != nil {
		log.Fatal(err)
	}
}

func Writef(w io.Writer, form string, args ...interface{}) {
	Writeln(w, fmt.Sprintf(form, args...))
}

func WriteChannel(w io.Writer, c string, s string) {
	Writef(w, "PRIVMSG %s :%s", c, s)
}
