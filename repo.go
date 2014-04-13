package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// idx-file
type idxFile struct {
	indexpath    string
	packpath     string
	packversion  uint32
	offsetValues map[sha1]uint64
}

// A Repository is the base of all other actions. If you need to lookup a
// commit, tree or blob, you do it from here.
type Repository struct {
	Path       string
	indexfiles []*idxFile

	commitCache map[sha1]*Commit
}

// Open the repository at the given path.
func OpenRepository(path string) (*Repository, error) {
	repo := new(Repository)
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	repo.Path = path
	fm, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fm.IsDir() {
		return nil, errors.New(fmt.Sprintf("%q is not a directory."))
	}

	indexfiles, err := filepath.Glob(filepath.Join(path, "objects/pack/*idx"))
	if err != nil {
		return nil, err
	}
	repo.indexfiles = make([]*idxFile, len(indexfiles))
	for i, indexfile := range indexfiles {
		idx, err := readIdxFile(indexfile)
		if err != nil {
			return nil, err
		}
		repo.indexfiles[i] = idx
	}

	return repo, nil
}
