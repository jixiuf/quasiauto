package quasiauto

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Sequence_Exec(t *testing.T) {
	parseKeyMap()
	empty := []int32{}
	args := map[string]string{"USERNAME": "user", "PASSWORD": "passw", "FIELD2": "=.2#'=<&):[.="}
	tests := []struct {
		name       string
		seqEntries SeqEntries
		wantTyped  []string
		wantTapped []string
		returnIds  []int32
	}{
		{"Field", []SeqEntry{{"USERNAME", nil, FIELD}}, []string{"user"}, nil, empty},
		{"Basic", []SeqEntry{
			{"USERNAME", nil, FIELD},
			{"TAB", nil, KEYWORD},
			{"PASSWORD", nil, FIELD},
			{"ENTER", nil, KEYWORD},
		}, []string{"user", "passw"}, []string{"tab", "enter"}, empty},
		{"Command", []SeqEntry{{"APPACTIVATE", []string{"foo"}, COMMAND}}, nil, nil, []int32{1}},
		{"Raw", []SeqEntry{{"hello dolly", nil, RAW}}, []string{"hello dolly"}, nil, empty},
		{"Special", []SeqEntry{
			{"^", nil, RAW},
			{"PASSWORD", nil, FIELD},
			{"%", nil, KEYWORD},
		}, []string{"passw", "%"}, []string{"ctrl"}, empty},
		{"Plus", []SeqEntry{{"+", nil, KEYWORD}}, []string{"+"}, nil, empty},
		{"Percent", []SeqEntry{{"%", nil, KEYWORD}}, []string{"%"}, nil, empty},
		{"Caret", []SeqEntry{{"^", nil, KEYWORD}}, []string{"^"}, nil, empty},
		{"Tilde", []SeqEntry{{"~", nil, KEYWORD}}, []string{"~"}, nil, empty},
		{"LParen", []SeqEntry{{"(", nil, KEYWORD}}, []string{"("}, nil, empty},
		{"RParen", []SeqEntry{{")", nil, KEYWORD}}, []string{")"}, nil, empty},
		{"At", []SeqEntry{{"AT", nil, KEYWORD}}, []string{"@"}, nil, empty},
		{"TAB", []SeqEntry{{"TAB", nil, KEYWORD}}, nil, []string{"tab"}, empty},
		{"ENTER", []SeqEntry{{"ENTER", nil, KEYWORD}}, nil, []string{"enter"}, empty},
		{"SPACE", []SeqEntry{{"SPACE", nil, KEYWORD}}, nil, []string{"space"}, empty},
		{"{", []SeqEntry{{"{", nil, KEYWORD}}, []string{"{"}, nil, empty},
		{"}", []SeqEntry{{"}", nil, KEYWORD}}, []string{"}"}, nil, empty},
		{"F1", []SeqEntry{{"F1", nil, KEYWORD}}, nil, []string{"f1"}, empty},
		{"NUMPAD5", []SeqEntry{{"NUMPAD5", nil, KEYWORD}}, nil, []string{"numpad_5"}, empty},
		{"Single meta tap", []SeqEntry{{"^", nil, RAW}}, nil, []string{"ctrl"}, empty},
		{"Single modified tap ", []SeqEntry{{"^c", nil, RAW}}, nil, []string{"ctrl", "c"}, empty},
		{"Muliple modifiers", []SeqEntry{{"^+c", nil, RAW}}, nil, []string{"ctrl", "shift", "c"}, empty},
		{"Muliple meta taps", []SeqEntry{{"%@", nil, RAW}}, nil, []string{"alt", "cmd"}, empty},
		{"Muliple meta modified taps", []SeqEntry{{"^+c^%c", nil, RAW}}, nil, []string{"ctrl", "shift", "c", "ctrl", "alt", "c"}, empty},
		{"Complex raw", []SeqEntry{{"ab^+cdd", nil, RAW}}, []string{"ab", "dd"}, []string{"ctrl", "shift", "c"}, empty},
		{"Embedded enter", []SeqEntry{{"ab~cd", nil, RAW}}, []string{"ab", "cd"}, []string{"enter"}, empty},
		{"Modified enter", []SeqEntry{{"%~", nil, RAW}}, nil, []string{"alt", "enter"}, empty},
		{"Funny characters", []SeqEntry{{"FIELD2", nil, FIELD}}, []string{"=.2#'=<&):[.="}, nil, empty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := New()
			mt.ReturnIds = tt.returnIds
			s := NewSequence()
			s.SeqEntries = tt.seqEntries
			s.Exec(args, &mt)
			assert.Equal(t, tt.wantTyped, mt.Typed, "typed")
			assert.Equal(t, tt.wantTapped, mt.Tapped, "tapped")
			if len(tt.returnIds) > 0 {
				assert.Equal(t, tt.returnIds[0], mt.Activated)
			}
		})
	}
}

func Test_handleCommand_DELAY(t *testing.T) {
	tests := []struct {
		token   string
		args    []string
		wantErr bool
	}{
		{"DELAY", []string{"10"}, false},
		{"DELAY", []string{"x"}, true},
		{"DELAY", []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			mt := New()
			se := SeqEntry{Token: tt.token, Args: tt.args, Type: COMMAND}
			err := handleCommand(&mt, se)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_handleCommand_APPACT(t *testing.T) {
	tests := []struct {
		name string
		args []string
		rids []int32
		rerr error
	}{
		{"bad args", nil, []int32{0}, errors.New("")},
		{"no error", []string{"no error"}, []int32{1}, nil},
		{"error", []string{"error"}, []int32{2}, errors.New("blah")},
		{"activated", []string{"activated"}, []int32{3}, nil},
		{"empty IDs", []string{"activated"}, []int32{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := New()
			mt.ReturnIds = tt.rids
			mt.ReturnErr = tt.rerr
			se := SeqEntry{Token: "APPACTIVATE", Args: tt.args, Type: COMMAND}
			err := handleCommand(&mt, se)
			if tt.rerr != nil {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				if len(tt.rids) > 0 {
					assert.Equal(t, tt.rids[0], mt.Activated)
				}
			}
		})
	}
}

func Benchmark_Exec(b *testing.B) {
	input := "aa^+bcc%@dee^~ff~gg"
	s := NewSequence()
	s.Parse(input)
	mt := New()
	args := make(map[string]string)
	for i := 0; i < b.N; i++ {
		s.Exec(args, &mt)
	}
}

type MockTyper struct {
	Typed     []string
	Tapped    []string
	Activated int32
	ReturnIds []int32
	ReturnErr error
}

func New() MockTyper {
	return MockTyper{
		ReturnIds: make([]int32, 0),
	}
}
func (r *MockTyper) TypeStr(s string, lag int) {
	if r.Typed == nil {
		r.Typed = make([]string, 0)
	}
	r.Typed = append(r.Typed, s)
}
func (r *MockTyper) KeyTap(s string, args ...interface{}) {
	if r.Tapped == nil {
		r.Tapped = make([]string, 0)
	}
	if args != nil {
		for _, a := range args {
			if m, ok := a.(string); ok {
				r.Tapped = append(r.Tapped, m)
			}
		}
	}
	r.Tapped = append(r.Tapped, s)
}
func (r *MockTyper) ActivePID(s int32) {
	r.Activated = s
}
func (r *MockTyper) FindIds(s string) ([]int32, error) {
	return r.ReturnIds, r.ReturnErr
}
