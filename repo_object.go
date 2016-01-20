package git

import (
	"fmt"
	"os"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound ObjectID

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", ObjectID(id))
}

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
	case ObjectTag:
		return "tag"
	default:
		return "invalid"
	}
}

type Object struct {
	Type ObjectType
	Size int64
	Data []byte
}

// Given a SHA1, find the pack it is in and the offset, or return nil if not
// found.
func (repo *Repository) findObjectPack(id ObjectID) (*idxFile, uint64) {
	for _, indexfile := range repo.indexfiles {
		if offset, ok := indexfile.offsetValues[id]; ok {
			return indexfile, offset
		}
	}
	return nil, 0
}

func (repo *Repository) Object(id ObjectID) (*Object, error) {
	return repo.getRawObject(id, false)
}

func (repo *Repository) getRawObject(id ObjectID, metaOnly bool) (*Object, error) {
	o, err := readObjectFile(filepathFromSHA1(repo.Path, id.String()), metaOnly)
	if err == nil {
		return o, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	if pack, _ := repo.findObjectPack(id); pack == nil {
		return nil, ObjectNotFound(id)
	}
	pack, offset := repo.findObjectPack(id)
	o, err = readObjectBytes(pack.packpath, &repo.indexfiles, offset, metaOnly)
	if err != nil {
		return nil, err
	}
	return o, nil
}
