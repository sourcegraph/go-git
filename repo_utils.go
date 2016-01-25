package git

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func readIdxFile(path string) (*idxFile, error) {
	ifile := &idxFile{}
	ifile.indexpath = path
	ifile.packpath = path[0:len(path)-3] + "pack"
	idx, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(idx, []byte{255, 't', 'O', 'c'}) {
		return nil, errors.New("Not version 2 index file")
	}
	pos := 8
	var fanout [256]uint32
	for i := 0; i < 256; i++ {
		// TODO: use range
		fanout[i] = uint32(idx[pos])<<24 + uint32(idx[pos+1])<<16 + uint32(idx[pos+2])<<8 + uint32(idx[pos+3])
		pos += 4
	}
	numObjects := int(fanout[255])
	ids := make([]ObjectID, numObjects)

	for i := 0; i < numObjects; i++ {
		ids[i] = ObjectID(idx[pos : pos+20])
		pos = pos + 20
	}
	// skip crc32 and offsetValues4
	pos += 8 * numObjects

	excessLen := len(idx) - 258*4 - 28*numObjects - 40
	var offsetValues8 []uint64
	if excessLen > 0 {
		// We have an index table, so let's read it first
		offsetValues8 = make([]uint64, excessLen/8)
		for i := 0; i < excessLen/8; i++ {
			offsetValues8[i] = uint64(idx[pos])<<070 + uint64(idx[pos+1])<<060 + uint64(idx[pos+2])<<050 + uint64(idx[pos+3])<<040 + uint64(idx[pos+4])<<030 + uint64(idx[pos+5])<<020 + uint64(idx[pos+6])<<010 + uint64(idx[pos+7])
			pos = pos + 8
		}
	}
	ifile.offsetValues = make(map[ObjectID]uint64, numObjects)
	pos = 258*4 + 24*numObjects
	for i := 0; i < numObjects; i++ {
		offset := uint32(idx[pos])<<24 + uint32(idx[pos+1])<<16 + uint32(idx[pos+2])<<8 + uint32(idx[pos+3])
		offset32ndbit := offset & 0x80000000
		offset31bits := offset & 0x7FFFFFFF
		if offset32ndbit == 0x80000000 {
			// it's an index entry
			ifile.offsetValues[ids[i]] = offsetValues8[offset31bits]
		} else {
			ifile.offsetValues[ids[i]] = uint64(offset31bits)
		}
		pos = pos + 4
	}
	// ObjectIDPackfile := idx[pos : pos+20]
	// ObjectIDIndex := idx[pos+21 : pos+40]
	fi, err := os.Open(ifile.packpath)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	packVersion := make([]byte, 8)
	_, err = fi.Read(packVersion)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(packVersion, []byte{'P', 'A', 'C', 'K'}) {
		return nil, errors.New("Pack file does not start with 'PACK'")
	}
	ifile.packversion = uint32(packVersion[4])<<24 + uint32(packVersion[5])<<16 + uint32(packVersion[6])<<8 + uint32(packVersion[7])
	return ifile, nil
}

func filepathFromSHA1(rootdir, id string) string {
	return filepath.Join(rootdir, "objects", id[:2], id[2:])
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
