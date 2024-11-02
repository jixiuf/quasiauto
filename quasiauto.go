package quasiauto

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Pairs represents lines of key/value pairs parsed from the input.
type Pairs map[string]string

func Parse(in io.Reader) (Sequence, Pairs, error) {
	rvp := make(Pairs)
	s := bufio.NewScanner(in)
	s.Scan()
	rvs := NewSequence()
	err := rvs.Parse(s.Text())
	if err != nil {
		return rvs, rvp, err
	}
	for s.Scan() {
		l := s.Text()
		ks := strings.SplitN(l, "\t", 2)
		if len(ks) == 2 {
			rvp[strings.ToUpper(ks[0])] = ks[1]
		} else {
			return rvs, rvp, fmt.Errorf("Illegal line format; expected KEY\\tVALUE, but got %d parts\n", len(ks))
		}
	}
	return rvs, rvp, nil
}
