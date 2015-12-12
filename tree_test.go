package git

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func openRepo(path string, treeish string) (*Commit, error) {
	r, err := OpenRepository(path)
	if err != nil {
		return nil, err
	}
	ci, err := r.GetCommitOfBranch(treeish)
	if err == nil {
		return ci, err
	}
	ci, err = r.GetCommitOfTag(treeish)
	if err == nil {
		return ci, err
	}
	return r.GetCommit(treeish)
}

func listWalk(t Tree) ([]string, error) {
	paths := []string{}
	err := t.Walk(func(path string, te *TreeEntry, err error) error {
		if err != nil {
			return err
		}
		p := filepath.Join(path)
		if te.IsDir() {
			p += "/"
		}
		paths = append(paths, p)
		return nil
	})
	sort.Strings(paths)
	return paths, err
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

func TestListEntriesCommit(t *testing.T) {
	ci, err := openRepo("testdata/test3/.git", "85d3a39020cf28af4b887552fcab9e31a49f2ced")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := ci.ListEntries()
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	expect := []string{
		"file1",
		"link1",
	}
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("got %q; want %q", got, expect)
	}
}

func TestSubmodules(t *testing.T) {
	r, err := OpenRepository("testdata/submodules/.git")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Commit string
		Expect []*Submodule
	}{
		{
			Commit: "b0817c6a0ee0670a4f36e49b3d5bb90453359f12",
			Expect: []*Submodule{
				&Submodule{Name: "sub", Path: "sub", URL: "../test.git"},
				&Submodule{Name: "sub2", Path: "sub2", URL: "../test2"},
			},
		},
		{
			Commit: "e1a2bf02b14c3db9193f17b872d3ddf48318a64b",
			Expect: []*Submodule{
				&Submodule{Name: "sub", Path: "sub", URL: "../test.git"},
			},
		},
		{
			Commit: "b3e93567fb6f97988eceeb8af2629a11dc47eea3",
			Expect: nil,
		},
	}

	for i, test := range tests {
		ci, err := r.GetCommit(test.Commit)
		got, err := ci.GetSubmodules()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(got, test.Expect) {
			t.Errorf("case %d: got %q; want %q", i, got, test.Expect)
		}
	}
}

func submoduleNames(subs []*Submodule) []string {
	r := []string{}
	for _, s := range subs {
		r = append(r, s.Name)
	}
	return r
}
