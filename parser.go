package navytrap

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Parsed struct {
	Nick    string   // source nick (RFC 1459 <servername>)
	Uinf    string   // user info
	Cmd     string   // IRC command
	Channel string   // normalized channel name
	Raw     string   // raw message
	Args    []string // parsed message parameters
}

var globalCmds = make(map[string]struct{})

func isNumeric(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}

// parse returns a filled Parsed structure representing its input.
func Parse(input string) (*Parsed, error) {
	var p = &Parsed{}
	var hasPrefix bool
	p.Raw = string(input)

	// step over leading :
	if input[0] == ':' {
		hasPrefix = true
		input = input[1:]
	}

	// split on spaces, unless a trailing param is found
	splf := func(data []byte, atEOF bool) (advance int, token []byte,
		err error) {
		if data[0] == ':' && len(data) > 1 { // trailing
			return 0, data[1:], bufio.ErrFinalToken
		}

		if !bytes.ContainsRune(data, ' ') {
			return 0, data, bufio.ErrFinalToken
		}

		i := 0
		for ; i < len(data); i++ {
			if data[i] == ' ' {
				break
			}
		}
		return i + 1, data[:i], nil
	}
	in := bufio.NewScanner(strings.NewReader(input))
	in.Split(splf)

	// prefix
	if hasPrefix {
		if ok := in.Scan(); !ok {
			return p, fmt.Errorf("expected prefix")
		}
		if strings.Contains(in.Text(), "!") { // userinfo included
			pref := strings.Split(in.Text(), "!")
			p.Nick = pref[0]
			p.Uinf = pref[1]
		} else {
			p.Nick = in.Text()
		}
	}

	// command
	if ok := in.Scan(); !ok {
		return p, fmt.Errorf("expected command")
	}
	p.Cmd = in.Text()

	// params
	for i := 0; in.Scan(); i++ {
		p.Args = append(p.Args, in.Text())
	}

	// set channel of normal messages. numeric (server) replies and
	// non-channel-specific commands will have .channel = ""
	if _, ok := globalCmds[p.Cmd]; !ok && !isNumeric(p.Cmd) {
		p.Channel = strings.ToLower(p.Args[0])
	}
	return p, nil
}
