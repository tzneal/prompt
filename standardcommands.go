package prompt

import (
	"fmt"
	"io"
)

// PushCommandSet returns a command that enters a named command set.
func PushCommandSet(p *Prompt, name string) Command {
	return func(w io.Writer, args []string) {
		p.PushCommandSet(name)
	}
}

// PushCommandSetArg returns a command that enters the command set passed in
// as an argument. It should be registered with a description like "enter $1"
// to specify that it takes a single argument.
func PushCommandSetArg(p *Prompt) Command {
	return func(w io.Writer, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(w, "expected a single argument")
			return
		}
		p.PushCommandSet(args[0])
	}
}

// PopCommandSet returns a command that enters a leaves the current command set.
func PopCommandSet(p *Prompt) Command {
	return func(w io.Writer, args []string) {
		p.PopCommandSet()
	}
}
