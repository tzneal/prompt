package prompt

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func TestCommandSetStack(t *testing.T) {
	p := NewPrompt()
	defer p.Close()
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

func TestPromptCompleter(t *testing.T) {
	p := NewPrompt()
	defer p.Close()

	cs := p.NewCommandSet("foo")
	cs.RegisterCommand("test", func(io.Writer, []string) {})
	cs.RegisterCommand("testing", func(io.Writer, []string) {})
	cs.RegisterCommand("traffic", func(io.Writer, []string) {})

	tests := []struct {
		input string
		exp   []string
	}{{"t", []string{"test", "testing", "traffic"}},
		{"te", []string{"test", "testing"}},
		{"tes", []string{"test", "testing"}},
		{"test", []string{"testing"}},
		{"testb", nil}}
	for _, tc := range tests {
		got := p.inputCompleter(tc.input)
		if len(got) != len(tc.exp) {
			t.Fatalf("expected %v, got %v", tc.exp, got)
		}
		for i := range tc.exp {
			if tc.exp[i] != got[i] {
				t.Fatalf("expected %s, got %s", tc.exp[i], got[i])
			}
		}
	}
}

func buildTestPrompt(t *testing.T) (prompt *Prompt, cleanup func()) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	repStdin, err := ioutil.TempFile("", "prompt")
	if err != nil {
		t.Fatalf("unable to create temp file: %s", err)
	}
	os.Stdin = repStdin
	repStdout, err := ioutil.TempFile("", "prompt")
	if err != nil {
		t.Fatalf("unable to create temp file: %s", err)
	}
	os.Stdout = repStdout

	p := NewPrompt()
	return p, func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		p.Close()
	}
}

func TestPromptNoCommandsRegistered(t *testing.T) {
	p, cleanup := buildTestPrompt(t)
	defer cleanup()

	fmt.Fprintf(os.Stdin, "test\n")
	_, err := os.Stdin.Seek(0, 0)

	if err != nil {
		t.Fatalf("unable to seek file: %s", err)
	}
	if p.Prompt() != true {
		t.Errorf("expected Prompt() = true with input")
	}
	if p.Prompt() != false {
		t.Errorf("expected Prompt() = false with eof")
	}
}

func TestPromptSingleCommandNoArgs(t *testing.T) {
	p, cleanup := buildTestPrompt(t)
	defer cleanup()

	cmdExecuted := false
	cs := p.NewCommandSet("foo")
	cs.RegisterCommand("test", func(io.Writer, []string) {
		cmdExecuted = true
	})

	fmt.Fprintf(os.Stdin, "test\n")
	_, err := os.Stdin.Seek(0, 0)

	if err != nil {
		t.Fatalf("unable to seek file: %s", err)
	}
	if p.Prompt() != true {
		t.Errorf("expected Prompt() = true with input")
	}
	if p.Prompt() != false {
		t.Errorf("expected Prompt() = false with eof")
	}
	if cmdExecuted == false {
		t.Error("expected command to run")
	}
}

func TestPromptSingleCommandSingleArg(t *testing.T) {
	p, cleanup := buildTestPrompt(t)
	defer cleanup()

	cmdArg := ""
	cs := p.NewCommandSet("foo")
	cs.RegisterCommand("test $1", func(w io.Writer, args []string) {
		cmdArg = args[0]
	})

	fmt.Fprintf(os.Stdin, "test bar\n") // should run
	fmt.Fprintf(os.Stdin, "test\n")     // won't run
	_, err := os.Stdin.Seek(0, 0)

	if err != nil {
		t.Fatalf("unable to seek file: %s", err)
	}
	for p.Prompt() {

	}
	if cmdArg != "bar" {
		t.Error("expected command to run")
	}
}

func TestPromptSingleCommandWildcard(t *testing.T) {
	p, cleanup := buildTestPrompt(t)
	defer cleanup()

	cmdArgs := []string{}
	cs := p.NewCommandSet("foo")
	cs.RegisterCommand("test $*", func(w io.Writer, args []string) {
		cmdArgs = args
	})

	fmt.Fprintf(os.Stdin, "test bar baz\n")
	fmt.Fprintf(os.Stdin, "test\n")
	_, err := os.Stdin.Seek(0, 0)

	if err != nil {
		t.Fatalf("unable to seek file: %s", err)
	}

	for p.Prompt() {
	}

	if len(cmdArgs) != 2 {
		t.Fatalf("expected command to run")
	}
	if cmdArgs[0] != "bar" || cmdArgs[1] != "baz" {
		t.Errorf("expected args = bar, baz, got %v", cmdArgs)
	}
}

func randCmd(r rand.Source, maxLen int64) []byte {
	b := make([]byte, r.Int63()%maxLen+1)
	for i := range b {
		c := byte(r.Int63() % 255)
		// prevent file output
		if c == '>' {
			c = ' '
		}
		b[i] = c
	}
	b[len(b)-1] = '\n'
	return b
}

func TestPromptFuzz(t *testing.T) {
	p, cleanup := buildTestPrompt(t)
	defer cleanup()

	cs := p.NewCommandSet("foo")
	cs.RegisterCommand("test", func(w io.Writer, args []string) {
	})

	r := rand.NewSource(42)
	for i := 0; i < 25000; i++ {
		os.Stdin.Write(randCmd(r, 40))
	}
	_, err := os.Stdin.Seek(0, 0)

	if err != nil {
		t.Fatalf("unable to seek file: %s", err)
	}
	for p.Prompt() {

	}
	// no tests, just insuring garbage input won't crash
}
