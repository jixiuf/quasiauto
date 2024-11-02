package quasiauto

import (
	"fmt"
	"github.com/gen2brain/beeep"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Exec uses key/value pairs to interpret and process a Sequence
// The keys of the pairs are expected to be normalized to upper case.
func (s Sequence) Exec(ds Pairs, typer Typer) {
	// TODO Is there a better way of ensuring no meta keys are pressed before we start typing?
	// Give use time to release the control key
	var err error
	metas := regexp.MustCompile("([+^%@]+[a-z~]?)|([^+^%@~]+)|(~)")
	lag := time.Duration(s.Keylag)
	// Give a little pause before we start doing anything
	time.Sleep(lag)
	for _, seq := range s.SeqEntries {
		err = nil
		switch seq.Type {
		case FIELD:
			v := ds[strings.ToUpper(seq.Token)]
			typer.TypeStr(v, s.Keylag)
		case KEYWORD:
			i := keyMap[seq.Token]
			if i.isTap {
				typer.KeyTap(i.val)
				time.Sleep(lag)
			} else {
				typer.TypeStr(i.val, s.Keylag)
			}
		case COMMAND:
			err = handleCommand(typer, seq)
		case RAW:
			parts := metas.FindAllString(seq.Token, -1)
			for _, part := range parts {
				if strings.ContainsAny(part, "+^%@~") {
					var typing string
					mods := make([]interface{}, 0)
					for i := 0; i < len(part); i++ {
						switch part[i] {
						case '+':
							mods = append(mods, "shift")
						case '^':
							mods = append(mods, "ctrl")
						case '%':
							mods = append(mods, "alt")
						case '@':
							mods = append(mods, "cmd")
						case '~':
							typing = "enter"
						default:
							typing = string([]byte{part[i]})
						}
					}
					// If it's just a sequence of modifier keys, just tap them out
					if len(typing) == 0 {
						for _, m := range mods {
							if str, ok := m.(string); ok {
								typer.KeyTap(str)
								time.Sleep(lag)
							}
						}
					} else if len(mods) == 0 {
						typer.KeyTap(typing)
						time.Sleep(lag)
					} else {
						typer.KeyTap(typing, mods...)
					}
				} else {
					typer.TypeStr(part, s.Keylag)
				}
			}
		default:
			log.Printf("unknown sequence type %d", seq.Type)
		}
		if err != nil {
			log.Printf("Sequence.exec(): %s", err)
		}
	}
}

// Typer is a UI interface; it outputs data and minimally interacts with
// windows. It's mainly an interface to allow mocking robotgo, which implements
// all functions as top-level package functions.
type Typer interface {
	// TypeStr outputs a string to the focused window
	TypeStr(string, int)
	// KeyTap takes keycode descriptions, like "enter" and "tab" and outputs
	// character codes
	KeyTap(string, ...interface{})
	// FindIds finds window IDs for the named application
	FindIds(string) ([]int32, error)
	// ActivePID makes the window the ID active
	ActivePID(int32)
}

// handleCommand processes tokens with arguments, such as DELAY and VKEY
func handleCommand(t Typer, seq SeqEntry) error {
	if len(seq.Args) < 1 {
		return fmt.Errorf("expected arguments for token %s", seq.Token)
	}
	switch seq.Token {
	case "DELAY":
		if len(seq.Args) != 1 {
			return fmt.Errorf("DELAY takes 1 integer argument")
		}
		d, e := strconv.Atoi(seq.Args[0])
		if e != nil {
			return fmt.Errorf("bad argument %s for DELAY: %s", seq.Args[0], e)
		}
		time.Sleep(time.Duration(d) * time.Millisecond)
	case "VKEY":
		// FIXME implement VKEY
		return fmt.Errorf("VKEY is not implemented")
	case "APPACTIVATE":
		ps, err := t.FindIds(seq.Args[0])
		if err != nil {
			return err
		}
		if len(ps) > 0 {
			t.ActivePID(ps[0])
		}
	case "BEEP":
		if len(seq.Args) != 2 {
			return fmt.Errorf("BEEP takes two arguments, frequency (Hz) & duration (ms)")
		}
		freq, e := strconv.ParseFloat(seq.Args[0], 64)
		if e != nil {
			return fmt.Errorf("bad frequency argument %s for BEEP: %s", seq.Args[0], e)
		}
		duration, e := strconv.Atoi(seq.Args[1])
		if e != nil {
			return fmt.Errorf("bad duration argument %s for BEEP: %s", seq.Args[0], e)
		}
		beeep.Beep(freq, duration)
	}
	return nil
}
