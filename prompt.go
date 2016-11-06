package prompt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/peterh/liner"
)

// Prompt is the user prompt.
type Prompt struct {
	LineState   *liner.State           // the liner used for input, visibile to allow direct manipulation/changes
	Prompter    func() string          // Prompt is the function called to return the prompt
	curPrompt   string                 // the current prompt passed to liner
	commandSets map[string]*CommandSet // registered command sets
	completers  map[string]Completer   // context-sensitive placeholder completion
	filters     map[string]Filter      // filtering of command output
	cmdSetStack []*CommandSet          // stack of command sets that have been pushed
}

// NewPrompt returns a newly initialized prompt.
func NewPrompt() *Prompt {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	line.SetTabCompletionStyle(liner.TabPrints)
	p := &Prompt{
		Prompter: func() string {
			return "> "
		},
		LineState:   line,
		completers:  map[string]Completer{},
		filters:     map[string]Filter{},
		commandSets: map[string]*CommandSet{},
	}
	line.SetCompleter(p.inputCompleter)
	return p
}

// Close closes and cleans up the prompt.
func (p *Prompt) Close() error {
	err := p.LineState.Close()
	p.LineState = nil
	return err
}

// NewCommandSet constructs and registers a new command set.
func (p *Prompt) NewCommandSet(name string) *CommandSet {
	cs := newCommandSet(name)
	p.commandSets[name] = cs
	if p.cmdSetStack == nil {
		p.cmdSetStack = append(p.cmdSetStack, cs)
	}
	return cs
}

// extractArgs takes user input and a command description and returns the input
// arguments parsed against the command description, including reordering if
// necessary.
func extractArgs(input, cmd []segment) []string {
	args := []string{}
	indices := []int{}
	for i := range input {
		cw := cmd[i]
		if cw.typ == placeholderType {
			if cw.value == "*" {
				// a $* consumes the rest of the arugments, so copy the remainder
				for _, w := range input[i:] {
					args = append(args, w.value)
				}
				break
			} else {
				args = append(args, input[i].value)
				idx, _ := strconv.Atoi(cw.value)
				indices = append(indices, idx)
			}
		}
	}
	// sort our argument list by the placeholder indices
	sort.Sort(argSort{indices, args})
	return args
}

// argSort is used for sorting input arguments by placeholder indices
type argSort struct {
	indices []int
	args    []string
}

func (s argSort) Len() int           { return len(s.indices) }
func (s argSort) Less(i, j int) bool { return s.indices[i] < s.indices[j] }
func (s argSort) Swap(i, j int) {
	s.indices[i], s.indices[j] = s.indices[j], s.indices[i]
	s.args[i], s.args[j] = s.args[j], s.args[i]
}

func (p *Prompt) runCommand(match *command, input input) {
	var out io.Writer = os.Stdout
	close := []io.Closer{}
	// redirecting to a file?
	if input.outputFile != "" {
		f, err := os.Create(input.outputFile)
		if err != nil {
			fmt.Printf("error writing: %s\n", err)
		}
		out = f
		close = append(close, f)
	}

	// apply filters to output
	for i := range input.filters {
		filter := input.filters[len(input.filters)-i-1]
		if fc, ok := p.filters[filter.cmd]; !ok {
			fmt.Printf("%s is not a valid filter\n", filter.cmd)
			return
		} else {
			pr, pw := io.Pipe()
			close = append(close, pr)
			go fc(pr, out, filter.args)
			out = pw
		}
	}
	match.execute(out, extractArgs(input.words, match.desc.words))
	for _, c := range close {
		c.Close()
	}
}

// Prompt prompts the user and returns input.
func (p *Prompt) Prompt() bool {
	if userInput, err := p.LineState.Prompt(p.Prompter()); err == nil {
		// user just hit enter with no input
		if len(userInput) == 0 {
			return true
		}

		parsed, err := parse(userInput)
		if err != nil {
			fmt.Printf("parse error: %s\n", err)
			return true
		}

		for _, input := range parsed {
			match := p.execMatch(input)
			if match != nil {
				p.runCommand(match, input)
				p.LineState.AppendHistory(input.asUser())
			} else {
				fmt.Printf("%s: command not found\n", input.asUser())
				return true
			}
		}
		return true
	}

	return false
}

// RegisterCompleter registers a function to be used for context sensitive
// completion of command placeholders.
func (p *Prompt) RegisterCompleter(name string, fn Completer) error {
	if _, ok := p.completers[name]; ok {
		return fmt.Errorf("%s is already registered", name)
	}
	p.completers[name] = fn
	return nil
}

// PopCommandSet removes the latest command set pushed.
func (p *Prompt) PopCommandSet() error {
	if len(p.cmdSetStack) > 1 {
		p.cmdSetStack = p.cmdSetStack[0 : len(p.cmdSetStack)-1]
		return nil
	}
	return errors.New("can't pop command set")
}

// PushCommandSet enters a new command set and pushes it to the stack.
func (p *Prompt) PushCommandSet(name string) error {
	if cs, ok := p.commandSets[name]; ok {
		p.cmdSetStack = append(p.cmdSetStack, cs)
		return nil
	}
	return fmt.Errorf("unknown command set: %s", name)
}

// CurrentCommandSet returns the current command set in use.
func (p *Prompt) CurrentCommandSet() *CommandSet {
	if len(p.cmdSetStack) == 0 {
		return nil
	}
	return p.cmdSetStack[len(p.cmdSetStack)-1]
}

// RegisterFilter registers a filter for use by the user.
// In the case:
//    cmd | foo arg1 arg2
// 'foo' is the filter name,  and []string{"arg1","arg2"} would be
// passed to the filter.
func (p *Prompt) RegisterFilter(name string, fn Filter) error {
	if _, ok := p.filters[name]; ok {
		return fmt.Errorf("filter %s is already registered", name)
	}
	p.filters[name] = fn
	return nil
}

func (p *Prompt) execMatch(inp input) *command {
	if p.CurrentCommandSet() == nil {
		return nil
	}
	var cmdMatch *command
	var cmdMatchScore matchType
	for _, cmd := range p.CurrentCommandSet().commands {
		if mt := cmd.execMatch(inp); mt != matchNone && mt > cmdMatchScore {
			cmdMatch = cmd
			cmdMatchScore = mt
		}
	}
	return cmdMatch
}

func (p *Prompt) inputCompleter(line string) []string {
	l, err := parse(line)
	if err != nil {
		// TODO: notify user of an error
		return nil
	}

	type cmatch struct {
		mt          completionType
		completions []string
	}
	// complete command words
	cMatches := []cmatch{}
	var lastInput []segment
	if len(l) > 0 {
		lastInput = l[len(l)-1].words
	}

	hasPartialMatches := false
	hasExactMatches := false
	for _, cmd := range p.CurrentCommandSet().commands {
		mt, completions := cmd.complete(lastInput, p.completers)
		if mt == completeExact {
			hasExactMatches = true
		}
		if mt == completePartial {
			hasPartialMatches = true
		}

		if mt != completeNone {
			cMatches = append(cMatches, cmatch{mt, completions})
		}
	}

	// no completion matches
	if len(cMatches) == 0 {
		return nil
	}

	// partial match is completing a term the user is currently typing: fo -> foo
	// exact match means that the user has exactly matched a command segnment,
	// and we should complete with next segment: foo -> foo bar

	// if we have both sorts of matches, drop the exact matches and only complete
	// the partial word matches
	if hasPartialMatches && hasExactMatches {
		for i, cm := range cMatches {
			if cm.mt == completeExact {
				cMatches[i].mt = completeNone
			}
		}
		hasExactMatches = false
	}

	cWords := []string{}
	// form our list of completions
	for _, cm := range cMatches {
		if cm.mt == completeNone {
			continue
		}
		cWords = append(cWords, cm.completions...)
	}

	// remove completion duplicates by sorting and matching against
	// the previous word
	sort.Strings(cWords)
	pw := ""

	// we are matching the next word, so we leave the user's input alone and add
	// a new segment to fill in with completions
	if hasExactMatches {
		lastInput = append(lastInput, segment{})
	}

	ret := []string{}
	for _, cw := range cWords {
		// empty or duplicate?
		if len(cw) == 0 || cw == pw {
			continue
		}
		pw = cw

		// replace the last segment of the user input with our completion
		lastInput[len(lastInput)-1].typ = wordType
		lastInput[len(lastInput)-1].value = cw
		// and add the completion to our list
		ret = append(ret, asUser(lastInput))
	}
	return ret
}

func asUser(inp []segment) string {
	b := bytes.Buffer{}
	for j, in := range inp {
		b.WriteString(in.value)
		if j+1 != len(inp) {
			b.WriteRune(' ')
		}
	}
	return b.String()
}
