package git

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// ObjectNotFound error returned when a repo query is performed for an ID that does not exist.
type ObjectNotFound ObjectID

func (id ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found: %s", ObjectID(id))
}

type ObjectType int

const (
	ObjectCommit   ObjectType = 0x10
	ObjectTree                = 0x20
	ObjectBlob                = 0x30
	ObjectTag                 = 0x40
	objectOfsDelta            = 0x60
	objectRefDelta            = 0x70
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
	Size uint64
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
	o, err := readLooseObject(filepathFromSHA1(repo.Path, id.String()), metaOnly)
	if err == nil {
		return o, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	pack, offset := repo.findObjectPack(id)
	if pack == nil {
		return nil, ObjectNotFound(id)
	}
	o, err = repo.readObjectFromPack(pack.packpath, offset, metaOnly)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func readLooseObject(path string, sizeonly bool) (*Object, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	br := bufio.NewReader(zr)

	typStr, err := br.ReadString(' ')
	if err != nil {
		return nil, err
	}
	var typ ObjectType
	switch typStr[:len(typStr)-1] {
	case "blob":
		typ = ObjectBlob
	case "tree":
		typ = ObjectTree
	case "commit":
		typ = ObjectCommit
	case "tag":
		typ = ObjectTag
	}

	sizeStr, err := br.ReadString(0)
	if err != nil {
		return nil, err
	}
	size, err := strconv.ParseUint(sizeStr[:len(sizeStr)-1], 10, 64)
	if err != nil {
		return nil, err
	}

	if sizeonly {
		return &Object{typ, size, nil}, nil
	}

	data, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, err
	}

	return &Object{typ, size, data}, nil
}
