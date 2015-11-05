package git

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func openRepo(path string, branch string) (*Commit, error) {
	r, err := OpenRepository(path)
	if err != nil {
		return nil, err
	}
	ci, err := r.GetCommitOfBranch(branch)
	if err != nil {
		return nil, err
	}
	return ci, nil
}

func TestEntryByPath(t *testing.T) {
	ci, err := openRepo("testdata/test2/.git", "master")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		want string
	}{
		{"bazdir/", "bazdir"},
		{"bazdir/quuxdir/", "quuxdir"},
		{"bazdir/baz", "baz"},
		{"bazdir/quuxdir/quux", "quux"},
		{"foo", "foo"},
		{"foolink", "foolink"},
	}
	for _, test := range tests {
		te, err := ci.GetTreeEntryByPath(test.path)
		if err != nil {
			t.Errorf("got err %q, want %q", err, test.want)
		}
		if te.Name() != test.want {
			t.Errorf("got %q, want %q", te.Name(), test.want)
		}
	}
}

func TestTreeEntries(t *testing.T) {
	ci, err := openRepo("testdata/test2/.git", "master")
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
