package quasiauto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewSequence(t *testing.T) {
	s := NewSequence()
	s.SeqEntries = append(s.SeqEntries, SeqEntry{})
	assert.Equal(t, 1, len(s.SeqEntries))
}

func Test_Parse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Parse", "{PLUS}{TAB}{F7}heyho{NUMPAD5}{DELAY 5}{VKEY 5 6}{APPACTIVATE window}{BEEP 100 200}{[}", 10},
		{"Token count", "{PLUS}{TAB}{F7}heyho{NUMPAD5}{TILDE}{ENTER}{[}{%}", 9},
		{"One token", "{PLUS}", 1},
		{"Empty", "", 0},
		{"Only a field", "{TOTP}", 1},
		{"Only a command", "{DELAY 5}", 1},
		{"Only text", "TOTP", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := NewSequence()
			seq.Parse(tt.input)
			if len(seq.SeqEntries) != tt.expected {
				t.Errorf("parseKeySeq() %s = %v, want %v", tt.name, seq.SeqEntries, tt.expected)
			}
		})
	}
}

func Test_ParseTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SeqEntry
	}{
		{"Keyword", "{TAB}", SeqEntry{Token: "TAB", Args: nil, Type: KEYWORD}},
		{"Keyword - left brace", "{{}", SeqEntry{Token: "{", Args: nil, Type: KEYWORD}},
		{"Keyword - right brace", "{}}", SeqEntry{Token: "}", Args: nil, Type: KEYWORD}},
		{"F-key", "{F11}", SeqEntry{Token: "F11", Args: nil, Type: KEYWORD}},
		{"Numpad", "{NUMPAD0}", SeqEntry{Token: "NUMPAD0", Args: nil, Type: KEYWORD}},
		{"Raw", "raw text", SeqEntry{Token: "raw text", Args: nil, Type: RAW}},
		{"Command", "{BEEP 300 123}", SeqEntry{Token: "BEEP", Args: []string{"300", "123"}, Type: COMMAND}},
		{"Character", "{~}", SeqEntry{Token: "~", Args: nil, Type: KEYWORD}},
		{"Special char", "@", SeqEntry{Token: "@", Args: nil, Type: RAW}},
		{"More raw", "^t", SeqEntry{Token: "^t", Args: nil, Type: RAW}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := NewSequence()
			seq.Parse(tt.input)
			assert.Equal(t, 1, len(seq.SeqEntries))
			if len(seq.SeqEntries) == 1 {
				e := seq.SeqEntries[0]
				assert.Equal(t, tt.expected.Token, e.Token)
				assert.Equal(t, tt.expected.Type, e.Type)
				assert.Equal(t, tt.expected.Args, e.Args)
			}
		})
	}
}

func Benchmark_Parse(b *testing.B) {
	input := "{BEEP}{PLUS}{USERNAME}{TAB}{F7}heyho{NUMPAD5}{TILDE}{PASSWORD}{ENTER}{[}{%}"
	for i := 0; i < b.N; i++ {
		seq := NewSequence()
		seq.Parse(input)
	}
}
