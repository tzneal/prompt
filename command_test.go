package prompt

import (
	"strings"
	"testing"
)

func TestWildcard(t *testing.T) {
	tests := []struct {
		input           string
		isWildcard      bool
		isSubstWildcard bool
	}{{"go test", false, false},
		{"go $1", false, true},
		{"go $1 $2", false, true},
		{"go $1 $*", true, true},
		{"go $*", true, false}}
	for _, tc := range tests {
		c, _ := parseCommand(tc.input, nil)
		if c.isWildcard() != tc.isWildcard {
			t.Errorf("expected %s isWildcard = %v", tc.input, tc.isWildcard)
		}
		if c.isSubstWildcard() != tc.isSubstWildcard {
			t.Errorf("expected %s isSubstWildcard = %v", tc.input, tc.isSubstWildcard)
		}
	}
}

func TestCommandExecMatchWildcard(t *testing.T) {
	cmd, err := parseCommand("go score $*", nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	tests := []struct {
		input    string
		expMatch matchType
	}{
		{"go", matchNone},
		{"go score", matchNone},
		{"go score a", matchWildcard},
		{"go score a b", matchWildcard},
		{"go score a b c", matchWildcard},
	}

	for _, tc := range tests {
		inp, _ := parse(tc.input)
		mt := cmd.execMatch(inp[0])
		if mt != tc.expMatch {
			t.Errorf("expected match=%v, got %v for %v", tc.expMatch, mt, tc.input)
		}
	}
}

func TestCommandExecMatchSubst(t *testing.T) {
	cmd, err := parseCommand("go score $1 $2", nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	tests := []struct {
		input    string
		expMatch matchType
	}{
		{"go", matchNone},
		{"go score", matchNone},
		{"go score a", matchNone},
		{"go score a b", matchSubstitution},
		{"go score a b c", matchNone},
		{"go foo", matchNone},
		{"go foo a", matchNone},
		{"go foo a b", matchNone},
	}

	for _, tc := range tests {
		inp, _ := parse(tc.input)
		mt := cmd.execMatch(inp[0])
		if mt != tc.expMatch {
			t.Errorf("expected match=%v, got %v for %v", tc.expMatch, mt, tc.input)
		}
	}
}

func TestCommandCompletePlaceholder(t *testing.T) {
	cmd, err := parseCommand("go score $1:test", nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	tests := []struct {
		input       string
		cType       completionType
		completions []string
	}{
		{"go", completePartial, []string{"score"}},
		{"go score", completeExact, []string{"a", "b", "c"}},
		{"go score foo", completeNone, nil},
		{"go score foo bar", completeNone, nil},
	}

	completers := map[string]Completer{}
	completers["test"] = func(s string) (r []string) {
		for _, p := range []string{"a", "b", "c"} {
			if strings.HasPrefix(p, s) {
				r = append(r, p)
			}
		}
		return r
	}

	for _, tc := range tests {
		inp, _ := parse(tc.input)
		cType, completions := cmd.complete(inp[0].words, completers)
		if cType != tc.cType {
			t.Errorf("expected ct=%s, got %s for %v", tc.cType, cType, tc.input)
		}
		if len(completions) != len(tc.completions) {
			t.Fatalf("expected %v = %v for %v", completions, tc.completions, tc.input)
		}
		for i := range completions {
			if completions[i] != tc.completions[i] {
				t.Errorf("expected %s = %s for %v", completions[i], tc.completions[i], tc.input)
			}
		}
	}
}

func TestCommandCompletePlaceWildcard(t *testing.T) {
	cmd, err := parseCommand("go score $*:test", nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	tests := []struct {
		input       string
		cType       completionType
		completions []string
	}{
		{"go", completePartial, []string{"score"}},
		{"go score", completeExact, []string{"a", "b", "c"}},
		{"go score a", completeNone, nil},
		{"go score a b", completeNone, nil},
		{"", completeExact, []string{"go"}},
	}

	completers := map[string]Completer{}
	completers["test"] = func(s string) (r []string) {
		for _, p := range []string{"a", "b", "c"} {
			if strings.HasPrefix(p, s) {
				r = append(r, p)
			}
		}
		return r
	}

	for _, tc := range tests {
		inp, _ := parse(tc.input)
		// handle empty string
		if inp == nil || len(inp) == 0 {
			inp = []input{{}}
		}
		cType, completions := cmd.complete(inp[0].words, completers)
		if cType != tc.cType {
			t.Errorf("expected ct=%s, got %s for %v", tc.cType, cType, tc.input)
		}
		if len(completions) != len(tc.completions) {
			t.Fatalf("expected %v = %v for %v", completions, tc.completions, tc.input)
		}
		for i := range completions {
			if completions[i] != tc.completions[i] {
				t.Errorf("expected %s = %s for %v", completions[i], tc.completions[i], tc.input)
			}
		}
	}
}
