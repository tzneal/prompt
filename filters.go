package prompt

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

// Filter is the type of function used for filtering command output.  They should
// synchronously read from r, apply any filtering logic and write
// the result to the w.
type Filter func(r io.Reader, w io.Writer, args []string)

var nl = []byte{'\n'}

// Grep is a very simple grep.
func Grep(r io.Reader, w io.Writer, args []string) {
	sc := bufio.NewScanner(r)

	// no filter
	if len(args) == 0 {
		io.Copy(w, r)
		return
	}

	re, err := regexp.Compile(args[0])
	if err != nil {
		fmt.Fprintf(w, "error compiling regexp: %s", err)
	}
	for sc.Scan() {
		line := sc.Text()
		if !re.MatchString(line) {
			continue
		}
		w.Write([]byte(line))
		w.Write(nl)
	}
}
