package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tzneal/prompt"
)

func main() {
	p := prompt.NewPrompt()
	p.RegisterCompleter("fileOrDir", prompt.CompleteFileOrDir)
	p.RegisterCompleter("name", func(name string) (matches []string) {
		for _, v := range []string{"Todd", "Elizabeth", "Ellen", "Eleanore"} {
			if strings.HasPrefix(v, name) {
				matches = append(matches, v)
			}
		}
		return
	})
	p.RegisterFilter("grep", prompt.Grep)

	cs := p.NewCommandSet("default")
	cs.RegisterCommand("exit", func(w io.Writer, args []string) {
		fmt.Fprintf(w, "exiting...\n")
		p.Close()
		os.Exit(0)
	})
	cs.RegisterCommand("names", prompt.PushCommandSet(p, "names-set"))
	cs.RegisterCommand("list-files", prompt.PushCommandSet(p, "list-files-set"))
	names := p.NewCommandSet("names-set")
	names.RegisterCommand("exit", prompt.PopCommandSet(p))
	names.RegisterCommand("hello $*:name", func(w io.Writer, names []string) {
		w.Write([]byte("Hello "))
		for i, name := range names {
			if i > 0 {
				if i+1 == len(names) {
					w.Write([]byte(" and "))
				} else {
					w.Write([]byte(", "))
				}
			}
			w.Write([]byte(name))
		}
		w.Write([]byte{'!', '\n'})
	})

	lf := p.NewCommandSet("list-files-set")
	lf.RegisterCommand("exit", prompt.PopCommandSet(p))
	lf.RegisterCommand("ls $1:fileOrDir", func(w io.Writer, args []string) {
		inp := args[0]
		if fi, err := os.Stat(inp); err != nil {
			fmt.Fprintf(w, "error listing %s: %s\n", inp, err)
		} else {
			if fi.IsDir() {
				fis, err := ioutil.ReadDir(inp)
				if err != nil {
					fmt.Fprintf(w, "error listing %s: %s\n", inp, err)
				}
				for _, fi := range fis {
					if fi.IsDir() {
						fmt.Fprintf(w, "<%s>\n", fi.Name())
					} else {
						fmt.Fprintf(w, "%s is %d bytes\n", fi.Name(), fi.Size())
					}
				}
			} else {
				fmt.Fprintf(w, "%s is %d bytes\n", fi.Name(), fi.Size())
			}
		}
	})

	defer p.Close()
	for p.Prompt() {

	}
}
