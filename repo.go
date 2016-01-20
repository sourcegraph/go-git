package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type idxFile struct {
	indexpath    string
	packpath     string
	packversion  uint32
	offsetValues map[ObjectID]uint64
}

type Repository struct {
	Path       string
	indexfiles map[string]*idxFile

	commitCache map[ObjectID]*Commit
	tagCache    map[ObjectID]*Tag
}

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
		return nil, errors.New(fmt.Sprintf("%q is not a directory.", fm.Name()))
	}

	indexfiles, err := filepath.Glob(filepath.Join(path, "objects/pack/*idx"))
	if err != nil {
		return nil, err
	}
	repo.indexfiles = make(map[string]*idxFile, len(indexfiles))
	for _, indexfile := range indexfiles {
		idx, err := readIdxFile(indexfile)
		if err != nil {
			return nil, err
		}
		repo.indexfiles[indexfile] = idx
	}

	return repo, nil
}
