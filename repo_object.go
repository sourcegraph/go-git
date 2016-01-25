package git

import (
	"bufio"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

func filepathFromSHA1(rootdir, id string) string {
	return filepath.Join(rootdir, "objects", id[:2], id[2:])
}

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

func (repo *Repository) readObjectFromPack(path string, offset uint64, sizeonly bool) (*Object, error) {
	offsetInt := int64(offset)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.Seek(offsetInt, os.SEEK_SET); err != nil {
		return nil, err
	}

	br := bufio.NewReader(file)

	x, err := binary.ReadUvarint(br)
	if err != nil {
		return nil, err
	}
	typ := ObjectType(x & 0x70)
	size := x&^0x7f>>3 + x&0xf

	switch typ {
	case ObjectCommit, ObjectTree, ObjectBlob, ObjectTag:
		if sizeonly {
			return &Object{typ, size, nil}, nil
		}

		data, err := readAndDecompress(br, size)
		if err != nil {
			return nil, err
		}
		return &Object{typ, size, data}, nil

	case objectOfsDelta, objectRefDelta:
		var base *Object
		switch typ {
		case objectOfsDelta:
			relOffset, err := readOffset(br)
			if err != nil {
				return nil, err
			}
			base, err = repo.readObjectFromPack(path, uint64(offsetInt-relOffset), false)
			if err != nil {
				return nil, err
			}

		case objectRefDelta:
			id := make([]byte, 20)
			if _, err := io.ReadFull(br, id); err != nil {
				return nil, err
			}
			base, err = repo.getRawObject(ObjectID(id), false)
			if err != nil {
				return nil, err
			}
		}

		d, err := readAndDecompress(br, size)
		if err != nil {
			return nil, err
		}

		_, n := binary.Uvarint(d) // length of base object
		d = d[n:]
		resultObjectLength, n := binary.Uvarint(d)
		d = d[n:]
		if sizeonly {
			return &Object{typ, resultObjectLength, nil}, nil
		}

		data, err := applyDelta(base.Data, d, resultObjectLength)
		if err != nil {
			return nil, err
		}
		return &Object{typ, resultObjectLength, data}, nil

	default:
		return nil, errors.New("unexpected type")
	}
}

func readOffset(r io.ByteReader) (int64, error) {
	var offset int64
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		offset |= int64(b & 0x7f)
		if b&0x80 == 0 {
			return offset, nil
		}
		offset = (offset + 1) << 7
	}
}

func readAndDecompress(r io.Reader, inflatedSize uint64) ([]byte, error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	buf := make([]byte, inflatedSize)
	if _, err := io.ReadFull(zr, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func applyDelta(base, delta []byte, resultLen uint64) ([]byte, error) {
	res := make([]byte, resultLen)
	insertPoint := res
	for {
		if len(delta) == 0 {
			return res, nil
		}
		opcode := delta[0]
		delta = delta[1:]

		if opcode&0x80 == 0 {
			// copy from delta
			copy(insertPoint, delta[:opcode])
			insertPoint = insertPoint[opcode:]
			delta = delta[opcode:]
			continue
		}

		// copy from base
		readNum := func(len uint) uint64 {
			var x uint64
			for i := uint(0); i < len; i++ {
				if opcode&(1<<i) != 0 {
					x |= uint64(delta[0]) << (i * 8)
					delta = delta[1:]
				}
			}
			return x
		}
		_ = readNum
		copyOffset := readNum(4)
		copyLength := readNum(3)
		if copyLength == 0 {
			copyLength = 1 << 16
		}
		copy(insertPoint, base[copyOffset:copyOffset+copyLength])
		insertPoint = insertPoint[copyLength:]
	}
}
