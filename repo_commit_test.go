package git

import "testing"

func Test_Repository_getCommitIdOfPackedRef(t *testing.T) {
	tests := []struct {
		testRepo    string
		refPath     string
		expCommitID string
		expError    error
	}{{
		testRepo: "repo2",
		refPath:  "refs/tags/master",
		expError: RefNotFound("refs/tags/master"),
	}, {
		testRepo:    "repo3",
		refPath:     "refs/tags/master",
		expCommitID: "8b61789a76de9edaa49b2529d3aaa302ba238c0b",
	}}
	for _, test := range tests {
		func() {
			r := openTestRepo(t, test.testRepo)
			commitID, err := r.getCommitIdOfPackedRef(test.refPath)
			if string(commitID) != test.expCommitID {
				t.Errorf("expected commit %q, got %q", test.expCommitID, string(commitID))
			}
			if err != test.expError {
				t.Errorf("expected error %q, got %q", test.expError, err)
			}
		}()
	}
}
