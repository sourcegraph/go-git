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
	for _, e := range ci.ListEntries() {
		rc, err := e.Blob().Data()
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(ioutil.Discard, rc)
		rc.Close()
	}
}
