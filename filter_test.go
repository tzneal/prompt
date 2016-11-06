package prompt_test

import (
	"bytes"
	"testing"

	"github.com/tzneal/prompt"
)

func TestGrep(t *testing.T) {
	tests := []struct {
		args  []string
		input string
		exp   string
	}{{[]string{}, "foo\n", "foo\n"},
		{[]string{"foo"}, "foo\n", "foo\n"},
		{[]string{"foo"}, "foo\nbar\n", "foo\n"},
		{[]string{"foo"}, "Foo\nbar\n", ""},
		{[]string{"-i", "foo"}, "Foo\nbar\n", "Foo\n"},
		{[]string{"foo", "-i"}, "Foo\nbar\n", "Foo\n"},
		{[]string{"-v", "foo"}, "foo\nbar\n", "bar\n"},
		{[]string{"-v", "foo"}, "a\nb\nfoo\nbar\n", "a\nb\nbar\n"},
		{[]string{"fo+"}, "foo\nbar\n", "foo\n"},
		{[]string{"fo++"}, "foo\nbar\n", "error compiling regexp: error parsing regexp: invalid nested repetition operator: `++`"},
	}
	for _, tc := range tests {
		rd := bytes.NewBufferString(tc.input)
		w := &bytes.Buffer{}
		prompt.Grep(rd, w, tc.args)
		if got := string(w.Bytes()); got != tc.exp {
			t.Errorf("expected %s, got %s", tc.exp, got)
		}
	}
}
