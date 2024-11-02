package quasiauto

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	// Field names, e.g. {USERNAME}
	FIELD = iota
	// Keywords, e.g. {TAB} or {ENTER}
	KEYWORD
	// Commands, e.g. {DELAY 5}
	COMMAND
	// Raw text, e.g. text not enclosed in {}
	RAW
)

// SeqEntry is a single token in a key sequence, denoted by the token, the
// parsed type, and any args if it is a command.
type SeqEntry struct {
	// Token is the processed text, stripped of {}
	Token string
	// Args is only set for COMMANDs, and will be nil otherwise
	Args []string
	// The type of the sequence entry, e.g. KEYWORD, COMMAND, etc.
	Type int
}

type SeqEntries []SeqEntry

// Sequence is a parsed sequence of tokens
type Sequence struct {
	SeqEntries
	Keylag int
}

// NewSequence returns a new Sequence instance bound to a typer. Unless mocking, this
// should normally be:
// ```
// s := NewSequence(Robot{})
// ```
func NewSequence() Sequence {
	if len(keyMap) == 0 {
		parseKeyMap()
	}
	return Sequence{
		make(SeqEntries, 0),
		50,
	}
}

// Parse processes a [Keepass autotype sequence](https://keepass.info/help/base/autotype.html)
// and returns the parsed keys in the order in which they occurred.
func (rv *Sequence) Parse(keySeq string) error {
	if len(keySeq) == 0 {
		return fmt.Errorf("received empty sequence")
	}
	matches := regexp.MustCompile("\\{[^\\}]+\\}|[^\\{\\}]+|\\s+|\\{\\{\\}|\\{\\}\\}").FindAllString(keySeq, -1)
	if len(matches) == 0 {
		return fmt.Errorf("received malformed sequence %#v", keySeq)
	}
	for _, match := range matches {
		var s SeqEntry
		switch {
		case match == "{{}":
			s.Token = "{"
			s.Type = KEYWORD
		case match == "{}}":
			s.Token = "}"
			s.Type = KEYWORD
		case match[0] == '{':
			match = strings.Trim(match, "{}")
			match = strings.Trim(match, " ")
			match = strings.Replace(match, "=", " ", 1)
			if len(match) == 0 {
				return fmt.Errorf("invalid key sequence {}")
			}
			parts := strings.Split(match, " ")
			s.Token = parts[0]
			t, found := keyMap[s.Token]
			if !found {
				s.Token = match
				s.Type = FIELD
			} else if t.isCommand {
				s.Type = COMMAND
				if len(parts) > 1 {
					s.Args = parts[1:]
				} else {
					s.Args = []string{}
				}
			} else {
				s.Token = match
				s.Type = KEYWORD
			}
		default:
			s.Token = match
			s.Type = RAW
		}
		rv.SeqEntries = append(rv.SeqEntries, s)
	}
	return nil
}

//go:embed map.csv
var keyMapRaw string
var keyMap map[string]Type

type Type struct {
	val       string
	isTap     bool
	isCommand bool
}

func parseKeyMap() {
	keyMap = make(map[string]Type)
	r := csv.NewReader(strings.NewReader(keyMapRaw))
	r.LazyQuotes = true
	r.TrailingComma = true
	r.Read() // Get rid of header
	for rs, err := r.Read(); !errors.Is(err, io.EOF); rs, err = r.Read() {
		keyMap[rs[0]] = Type{rs[2], rs[1] == "Tap", rs[1] == "Command"}
	}
}
