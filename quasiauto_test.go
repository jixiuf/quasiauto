package quasiauto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parse(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantSeq   SeqEntries
		wantPairs Pairs
		wantError bool
	}{
		{
			"Basic",
			"{USERNAME}{TAB}{PASSWORD}{ENTER}\nUserName\tuser\nPassword\tpass",
			SeqEntries{
				SeqEntry{"USERNAME", nil, FIELD},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"PASSWORD", nil, FIELD},
				SeqEntry{"ENTER", nil, KEYWORD},
			},
			Pairs{"USERNAME": "user", "PASSWORD": "pass"}, false,
		},
		{
			"TOTP",
			"{USERNAME}{TAB}{PASSWORD}{TAB}{DELAY 500}{TOTP}{ENTER}\nUserName\tuser\nPassword\tpass\nTOTP\t123456",
			SeqEntries{
				SeqEntry{"USERNAME", nil, FIELD},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"PASSWORD", nil, FIELD},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"DELAY", []string{"500"}, COMMAND},
				SeqEntry{"TOTP", nil, FIELD},
				SeqEntry{"ENTER", nil, KEYWORD},
			},
			Pairs{"USERNAME": "user", "PASSWORD": "pass", "TOTP": "123456"}, false,
		},
		{
			"Ending CR",
			"{USERNAME}{TAB}{PASSWORD}{ENTER}\nUserName\tuser\nPassword\tpass\n",
			SeqEntries{
				SeqEntry{"USERNAME", nil, FIELD},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"PASSWORD", nil, FIELD},
				SeqEntry{"ENTER", nil, KEYWORD},
			},
			Pairs{"USERNAME": "user", "PASSWORD": "pass"}, false,
		},
		{
			"Some text",
			"IAmLegend{TAB}{PASSWORD}{ENTER}\nPassword\tpass\n",
			SeqEntries{
				SeqEntry{"IAmLegend", nil, RAW},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"PASSWORD", nil, FIELD},
				SeqEntry{"ENTER", nil, KEYWORD},
			},
			Pairs{"PASSWORD": "pass"}, false,
		},
		{
			"Backslash",
			"{USERNAME}{TAB}{PASSWORD}{ENTER}\nUserName\tuser\nPassword\tab\\cd\n",
			SeqEntries{
				SeqEntry{"USERNAME", nil, FIELD},
				SeqEntry{"TAB", nil, KEYWORD},
				SeqEntry{"PASSWORD", nil, FIELD},
				SeqEntry{"ENTER", nil, KEYWORD},
			},
			Pairs{"USERNAME": "user", "PASSWORD": "ab\\cd"}, false,
		},
		{"Parse error - no keyword", "{}", SeqEntries{}, Pairs{}, true},
		{"Parse error - blank lines", "{USERNAME}\n\n\n\n", SeqEntries{}, Pairs{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.source)
			gotSeq, gotPairs, err := Parse(reader)
			if tt.wantError {
				assert.Error(t, err, "expected error, but got %#v, %#v", gotSeq, gotPairs)
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, tt.wantSeq, gotSeq.SeqEntries)
				assert.EqualValues(t, tt.wantPairs, gotPairs)
			}
		})
	}
}
