package git

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"

	"testing"
)

func TestBlob(t *testing.T) {
	r, err := OpenRepository("testdata/test.git")
	if err != nil {
		t.Fatal(err)
	}
	ci, err := r.GetCommitOfBranch("master")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ci.ListEntries() {
		rc, err := e.Blob().Data()
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(ioutil.Discard, rc)
		rc.Close()
	}
}

func TestEntries(t *testing.T) {
	r, err := OpenRepository("testdata/test2/.git")
	if err != nil {
		t.Fatal(err)
	}
	ci, err := r.GetCommitOfBranch("master")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := listWalk(ci.Tree)
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{
		"bazdir/",
		"bazdir/baz",
		"bazdir/quuxdir/",
		"bazdir/quuxdir/quux",
		"foo",
		"foolink",
	}
	if !reflect.DeepEqual(entries, expect) {
		t.Errorf("got %q; want %q", entries, expect)
	}
}

func listWalk(t Tree) ([]string, error) {
	paths := []string{}
	err := t.Walk(func(path string, te *TreeEntry, err error) error {
		if err != nil {
			return err
		}
		p := filepath.Join(path, te.Name())
		if te.IsDir() {
			p += "/"
		}
		paths = append(paths, p)
		return nil
	})
	sort.Strings(paths)
	return paths, err
}
