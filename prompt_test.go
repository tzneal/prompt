package prompt

import "testing"

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
		inp, _ := parse(tc.input)
		cmd, _ := parse(tc.cmd)

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

func TestCommandSetStack(t *testing.T) {
	p := NewPrompt()
	csA := p.NewCommandSet("a")
	csB := p.NewCommandSet("b")
	csC := p.NewCommandSet("c")

	if p.CurrentCommandSet() != csA {
		t.Error("expected initial command set to be current")
	}
	if err := p.PushCommandSet("b"); err != nil {
		t.Errorf("push should succeed, got %s", err)
	}
	if p.CurrentCommandSet() != csB {
		t.Error("expected command set B to be current")
	}
	if err := p.PushCommandSet("c"); err != nil {
		t.Errorf("push should succeed, got %s", err)
	}
	if p.CurrentCommandSet() != csC {
		t.Error("expected command set C to be current")
	}

	// pop C
	if err := p.PopCommandSet(); err != nil {
		t.Errorf("pop should succeed, got %s", err)
	}
	if p.CurrentCommandSet() != csB {
		t.Error("expected command set B to be current")
	}

	// pop B
	if err := p.PopCommandSet(); err != nil {
		t.Errorf("pop should succeed, got %s", err)
	}
	if p.CurrentCommandSet() != csA {
		t.Error("expected command set A to be current")
	}

	// pop A, should fail
	if err := p.PopCommandSet(); err == nil {
		t.Error("pop should fail")
	}
}
