package git

import (
	"io"
	"io/ioutil"

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
	entries, err := ci.ListEntries()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		rc, err := e.Blob().Data()
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(ioutil.Discard, rc)
		rc.Close()
	}
}

func TestCommitIsAncestor(t *testing.T) {
	r, err := OpenRepository("testdata/ancestors/.git")
	if err != nil {
		t.Fatalf("failed to open repository: %s", err)
	}

	tests := []struct {
		ancestor, descendent string
		IsAncestor           bool
	}{
		{"858dec50e19d591dfab60a161296401fb57622e1", "0de1a577483e6e32aff64f68413b8e20d5306bae", true},  // c (master) -> a
		{"858dec50e19d591dfab60a161296401fb57622e1", "5059c8ccec10abd8dcac455afced0dd688989f35", true},  // c (master) -> b
		{"f571933a526cd287ecaeeb0bce514ac732b99938", "5059c8ccec10abd8dcac455afced0dd688989f35", false}, // d (fork) -> b
		{"f571933a526cd287ecaeeb0bce514ac732b99938", "0de1a577483e6e32aff64f68413b8e20d5306bae", true},  // d (fork) -> b
	}

	for _, test := range tests {
		commit, err := r.GetCommit(test.ancestor)
		if err != nil {
			t.Errorf("failed to load commit: %s", err)
			continue
		}
		if want, got := test.IsAncestor, commit.IsAncestor(test.descendent); want != got {
			t.Errorf("%q.IsAncestor(%q) -> %v != %v", test.ancestor, test.descendent, got, want)
		}
	}
}
