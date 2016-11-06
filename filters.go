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
	caseSensitive := true
	invertMatch := false
	regex := ""
	for _, arg := range args {
		switch arg {
		case "-i":
			caseSensitive = false
		case "-v":
			invertMatch = true
		default:
			regex = arg
		}
	}

	if !caseSensitive {
		regex = fmt.Sprintf("(?i)%s", regex)
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		fmt.Fprintf(w, "error compiling regexp: %s", err)
		return
	}

	for sc.Scan() {
		line := sc.Text()
		matchesRe := re.MatchString(line)
		if (!invertMatch && !matchesRe) || (invertMatch && matchesRe) {
			continue
		}
		w.Write([]byte(line))
		w.Write(nl)
	}
}
