package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-vgo/robotgo"
	"ser1.net/quasiauto"
)

var Version string = "development"

// TODO add notify-send entry title? https://github.com/keybase/go-notifier
// TODO quasi-type mode -- just type & tab through fields
// TODO add mouse mode -- click == type & advance
// FIXME If user holds down control key, autotype does damage.
func main() {
	ms := flag.Int("ms", 50, "microsecond delay between key presses")
	ttl := flag.Bool("title", false, "Return the currently active window title instead of autotyping")
	v := flag.Bool("version", false, "print version of quasiauto")
	flag.Parse()
	if *v {
		fmt.Println(Version)
		return
	}
	if *ttl {
		fmt.Printf(robotgo.GetTitle())
		return
	}
	// Read STDIN
	seq, kvs, err := quasiauto.Parse(os.Stdin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	seq.Keylag = *ms
	// Grab input
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		seq.Exec(kvs, Dotool{})
	} else {
		seq.Exec(kvs, Robot{})
	}

	os.Exit(0)
}

type Robot struct{}

func (r Robot) TypeStr(s string, lag int) {
	robotgo.TypeStr(s, float64(lag))
}
func (r Robot) KeyTap(s string, args ...interface{}) {
	robotgo.KeyTap(s)
}
func (r Robot) FindIds(s string) ([]int32, error) {
	return robotgo.FindIds(s)
}
func (r Robot) ActivePID(s int32) {
	robotgo.ActivePID(s)
}

type Dotool struct{}

func (r Dotool) TypeStr(s string, lag int) {
	// echo type $1|dotoolc
	run("sleep", fmt.Sprintf("%f", float64(lag)/1000))
	run("dotoolc", "type "+s)
}
func (r Dotool) KeyTap(s string, args ...interface{}) {
	run("sleep", "1")
	run("dotoolc", "key "+s)
}
func (r Dotool) FindIds(s string) ([]int32, error) {
	// return robotgo.FindIds(s)
	return nil, nil
}
func (r Dotool) ActivePID(s int32) {
	//  TODO:
	// robotgo.ActivePID(s)
}
func run(command, input string) error {
	var cmd *exec.Cmd
	cmd = exec.Command(command)

	// Set stdout to out var
	if input != "" {
		cmd.Stdin = bytes.NewBuffer([]byte(input))
	}

	// Run exec
	return cmd.Run()

}
