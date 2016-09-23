package prompt

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// Completer is used to provide context-sensitive completions for placeholders.
// Completers are registered, and then are called retrieve completions for
// the registered placeholder types.
type Completer func(arg string) []string

// CompleteFileOrDir is a pre-defined completer that is used to complete
// a file or directory name.
func CompleteFileOrDir(fileOrDir string) []string {
	// if the user supplied no input, pre-populate with the cwd
	if fileOrDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return []string{""}
		}
		fileOrDir = cwd
	}

	base := path.Clean(fileOrDir)
	dir, file := path.Split(base)
	// path specifies a directory, so complete inside of it
	if fs, err := os.Stat(fileOrDir); err == nil && fs.IsDir() {
		dir = fileOrDir
		file = ""
	}

	matchFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	cFiles := []string{}
	for _, mf := range matchFiles {
		if strings.HasPrefix(mf.Name(), file) {
			cf := path.Join(dir, mf.Name())
			if fs, err := os.Stat(cf); err == nil && fs.IsDir() {
				cFiles = append(cFiles, cf+string(os.PathSeparator))
			} else {
				cFiles = append(cFiles, cf)
			}
		}
	}
	return cFiles
}
