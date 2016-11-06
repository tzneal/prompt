package prompt

import "io"

// Command is a function representing a command to be executed.  Arguments are
// passed as args, and all output should be written to w to allow for
// filtering.
type Command func(w io.Writer, args []string)

// CommandSet is a set of commands, usually related.  Command sets can be
// switched between by registering commands that call PushCommandSet/PopCommandSet
// on the Prompt.
type CommandSet struct {
	name     string
	commands []*command
}

func newCommandSet(name string) *CommandSet {
	return &CommandSet{
		name: name,
	}
}

// RegisterCommandFunc registers a command. The description syntax is:
//
// cmd [arg1 arg2...] [$n:[completionType] ...] [$*:[completionType]]
//
// $n - n is a digit, and it matches a single argument. This can be used  to
// reorder arguments (e.g. "foo $2 $1", when called with "foo a b", will
// have []args{"b","a"}) when the command is executed.)
//
// $* - wildcard, matches all arguments to the end of the line
func (cs *CommandSet) RegisterCommandFunc(desc string, fn Command) error {
	cmd, err := parseCommand(desc, fn)
	if err != nil {
		return err
	}
	cs.commands = append(cs.commands, cmd)
	return nil
}
