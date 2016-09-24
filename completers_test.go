package prompt_test

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/tzneal/prompt"
)

func createTestDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}

	dirs := []string{"b1", "b2", "b31", "b32"}
	files := []string{"f.txt", "f.png", "f.tex"}
	for _, d := range dirs {
		os.MkdirAll(path.Join(dir, d), 0755)
	}
	for _, f := range files {
		ioutil.WriteFile(path.Join(dir, f), []byte{0x0}, 0644)
	}
	return dir
}

func TestCompleteFileOrDir(t *testing.T) {
	dir := createTestDir(t)
	defer os.RemoveAll(dir)

	tests := []struct {
		inp string
		exp []string
	}{
		{path.Join(dir, ""), []string{"/b1/", "/b2/", "/b31/", "/b32/", "/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/"), []string{"/b1/", "/b2/", "/b31/", "/b32/", "/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/b"), []string{"/b1/", "/b2/", "/b30/", "/b31/"}},
		{path.Join(dir, "/b3"), []string{"/b30/", "/b31/"}},
		{path.Join(dir, "/f"), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/f."), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/f.t"), []string{"/f.tex", "/f.txt"}},
		{path.Join(dir, "/f.te"), []string{"/f.tex"}},
		{path.Join(dir, "/f.tex"), []string{"/f.tex"}},
		{path.Join(dir, "/fasd"), nil},
	}
	baseDirLen := len(dir)
	for _, tc := range tests {
		res := prompt.CompleteFileOrDirectory(tc.inp)
		for i := range res {
			res[i] = res[i][baseDirLen:]
		}
		// get a consistent output
		sort.Strings(res)
		if len(res) != len(tc.exp) {
			t.Fatalf("expected %v = %v", res, tc.exp)
		}
	}

}

func TestCompleteDir(t *testing.T) {
	dir := createTestDir(t)
	defer os.RemoveAll(dir)

	tests := []struct {
		inp string
		exp []string
	}{
		{path.Join(dir, ""), []string{"/b1/", "/b2/", "/b31/", "/b32/"}},
		{path.Join(dir, "/"), []string{"/b1/", "/b2/", "/b31/", "/b32/"}},
		{path.Join(dir, "/b"), []string{"/b1/", "/b2/", "/b30/", "/b31/"}},
		{path.Join(dir, "/b3"), []string{"/b30/", "/b31/"}},
		{path.Join(dir, "/f"), nil},
		{path.Join(dir, "/f."), nil},
		{path.Join(dir, "/f.t"), nil},
		{path.Join(dir, "/f.te"), nil},
		{path.Join(dir, "/f.tex"), nil},
		{path.Join(dir, "/fasd"), nil},
	}
	baseDirLen := len(dir)
	for _, tc := range tests {
		res := prompt.CompleteDirectory(tc.inp)
		for i := range res {
			res[i] = res[i][baseDirLen:]
		}
		// get a consistent output
		sort.Strings(res)
		if len(res) != len(tc.exp) {
			t.Fatalf("expected %v = %v", res, tc.exp)
		}
	}

}

func TestCompleteFiles(t *testing.T) {
	dir := createTestDir(t)
	defer os.RemoveAll(dir)

	tests := []struct {
		inp string
		exp []string
	}{
		{path.Join(dir, ""), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/"), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/b"), nil},
		{path.Join(dir, "/b3"), nil},
		{path.Join(dir, "/f"), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/f."), []string{"/f.png", "/f.tex", "/f.txt"}},
		{path.Join(dir, "/f.t"), []string{"/f.tex", "/f.txt"}},
		{path.Join(dir, "/f.te"), []string{"/f.tex"}},
		{path.Join(dir, "/f.tex"), []string{"/f.tex"}},
		{path.Join(dir, "/fasd"), nil}}
	baseDirLen := len(dir)
	for _, tc := range tests {
		res := prompt.CompleteFile(tc.inp)
		for i := range res {
			res[i] = res[i][baseDirLen:]
		}
		// get a consistent output
		sort.Strings(res)
		if len(res) != len(tc.exp) {
			t.Fatalf("expected %v = %v", res, tc.exp)
		}
	}

}
