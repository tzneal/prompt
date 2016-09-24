package prompt

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		input string
		exp   string
		err   string
	}{{"foo", "[{<wordType foo>}]", ""},
		{"foo;bar;baz", "[{<wordType foo>} {<wordType bar>} {<wordType baz>}]", ""},
		{" foo; bar ; baz ", "[{<wordType foo>} {<wordType bar>} {<wordType baz>}]", ""},
		{"foo $*", "[{<wordType foo> <placeholderType *>}]", ""},
		{"foo $* $*", "[]", "duplicate placeholder $*"},
		{"foo $2 $1", "[{<wordType foo> <placeholderType 2> <placeholderType 1>}]", ""},
		{"foo $2 $2", "[]", "duplicate placeholder $2"},
		{"foo $*:host|grep -i 192 | grep -v 100>a.txt; ls a.txt",
			"[{<wordType foo> <placeholderType * complete:host> filter: grep[-i 192] grep[-v 100]} {<wordType ls> <wordType a.txt>}]", ""}}

	for _, tc := range testCases {
		inp, err := parseCmdDescription(tc.input)
		outp := fmt.Sprintf("%s", inp)
		if err != nil {
			if err.Error() != tc.err {
				t.Logf("TestCase: %s", tc.input)
				t.Errorf("expected err '%s', got '%s'", tc.err, err.Error())
			}
		}

		if err == nil && tc.err != "" {
			t.Logf("TestCase: %s", tc.input)
			t.Errorf("expected err %s, got no error", tc.err)
		}

		if tc.exp != outp {
			t.Logf("TestCase: %s", tc.input)
			t.Errorf("expected '%s', got '%s'", tc.exp, outp)
		}
	}
}

func TestArgExtraction(t *testing.T) {
	tests := []struct {
		cmd   string
		input string
		args  []string
	}{
		{"go $*", "go foo", []string{"foo"}},
		{"go $*", "go foo bar baz", []string{"foo", "bar", "baz"}},
		{"go $1", "go foo", []string{"foo"}},
		{"go $1 $2", "go foo bar", []string{"foo", "bar"}},
		{"go $2 $1", "go foo bar", []string{"bar", "foo"}},
		{"go $3 $1 $2", "go a b c", []string{"b", "c", "a"}},
	}

	for _, tc := range tests {
		inp, _ := parseUserInput(tc.input)
		cmd, _ := parseCmdDescription(tc.cmd)

		res := extractArgs(inp[0].words, cmd[0].words)
		if len(res) != len(tc.args) {
			t.Errorf("expected %s = %s", res, tc.args)
		}
		for i := range res {
			if res[i] != tc.args[i] {
				t.Errorf("expected arg %d, %s = %s", i, res[i], tc.args[i])
			}
		}
	}
}
