package git

import (
	"fmt"
	"io"
	"os"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound sha1

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", sha1(id))
}

// Who am I?
type ObjectType int

const (
	ObjectCommit ObjectType = 0x10
	ObjectTree   ObjectType = 0x20
	ObjectBlob   ObjectType = 0x30
	ObjectTag    ObjectType = 0x40
)

func (t ObjectType) String() string {
	switch t {
	case ObjectCommit:
		return "commit"
	case ObjectTree:
		return "tree"
	case ObjectBlob:
		return "blob"
	default:
		return ""
	}
}

// Given a SHA1, find the pack it is in and the offset, or return nil if not
// found.
func (repo *Repository) findObjectPack(id sha1) (*idxFile, uint64) {
	for _, indexfile := range repo.indexfiles {
		if offset, ok := indexfile.offsetValues[id]; ok {
			return indexfile, offset
		}
	}
	return nil, 0
}

func (repo *Repository) getRawObject(id sha1, metaOnly bool) (ObjectType, int64, io.ReadCloser, error) {
	ot, length, dataRc, err := readObjectFile(filepathFromSHA1(repo.Path, id.String()), metaOnly)
	if err == nil {
		return ot, length, dataRc, nil
	}
	if !os.IsNotExist(err) {
		return 0, 0, nil, err
	}

	if pack, _ := repo.findObjectPack(id); pack == nil {
		return 0, 0, nil, ObjectNotFound(id)
	}
	pack, offset := repo.findObjectPack(id)
	return readObjectBytes(pack.packpath, &repo.indexfiles, offset, metaOnly)
}

// Get (inflated) size of an object.
func (repo *Repository) objectSize(id sha1) (int64, error) {
	_, length, _, err := repo.getRawObject(id, true)
	return length, err
}
