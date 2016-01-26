package git

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type pack struct {
	repo      *Repository
	id        string
	indexFile *os.File
	packFile  *os.File
}

func (p *pack) indexFileReader() (io.ReaderAt, error) {
	if p.indexFile == nil {
		f, err := os.Open(filepath.Join(p.repo.Path, "objects", "pack", p.id+".idx"))
		if err != nil {
			return nil, err
		}
		p.indexFile = f
		if !bytes.Equal(readBytesAt(f, 0, 4), []byte{255, 't', 'O', 'c'}) {
			return nil, errors.New("wrong magic number")
		}
		if binary.BigEndian.Uint32(readBytesAt(f, 4, 4)) != 2 {
			return nil, errors.New("unsupported index file version")
		}
	}
	return p.indexFile, nil
}

func (p *pack) packFileReader() (io.ReaderAt, error) {
	if p.packFile == nil {
		f, err := os.Open(filepath.Join(p.repo.Path, "objects", "pack", p.id+".pack"))
		if err != nil {
			return nil, err
		}
		p.packFile = f
		if !bytes.HasPrefix(readBytesAt(f, 0, 4), []byte{'P', 'A', 'C', 'K'}) {
			return nil, errors.New("pack file does not min with 'PACK'")
		}
		if binary.BigEndian.Uint32(readBytesAt(f, 4, 4)) != 2 {
			return nil, errors.New("unsupported pack file version")
		}
	}
	return p.packFile, nil
}

func (p *pack) Close() error {
	if p.indexFile != nil {
		p.indexFile.Close()
	}
	if p.packFile != nil {
		p.packFile.Close()
	}
	return nil
}

func (p *pack) object(id ObjectID, metaOnly bool) (*Object, error) {
	r, err := p.indexFileReader()
	if err != nil {
		return nil, err
	}

	fanoutTableStart := int64(8)
	nameTableStart := fanoutTableStart + 4*256

	numObjects := binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*255, 4))

	checksumTableStart := nameTableStart + 20*int64(numObjects)
	offsetTableStart := checksumTableStart + 4*int64(numObjects)
	highOffsetTableStart := offsetTableStart + 4*int64(numObjects)
	_ = highOffsetTableStart

	firstByte := id[0]
	min := uint32(0)
	if firstByte > 0 {
		min = binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*int64(firstByte-1), 4))
	}
	max := binary.BigEndian.Uint32(readBytesAt(r, fanoutTableStart+4*int64(firstByte), 4)) - 1

	index, err := binarySearch(r, nameTableStart, min, max, id)
	if err != nil {
		return nil, err
	}

	offset := uint64(binary.BigEndian.Uint32(readBytesAt(r, offsetTableStart+4*int64(index), 4)))
	if offset&(1<<31) != 0 {
		highOffsetIndex := int64(offset &^ (1 << 31))
		offset = binary.BigEndian.Uint64(readBytesAt(r, highOffsetTableStart+8*highOffsetIndex, 8))
	}

	o, err := p.objectAtOffset(offset, metaOnly)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (p *pack) objectAtOffset(offset uint64, metaOnly bool) (*Object, error) {
	file, err := os.Open(filepath.Join(p.repo.Path, "objects", "pack", p.id+".pack"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.Seek(int64(offset), os.SEEK_SET); err != nil {
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
		if metaOnly {
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
			base, err = p.objectAtOffset(offset-relOffset, false)
			if err != nil {
				return nil, err
			}

		case objectRefDelta:
			id := make([]byte, 20)
			if _, err := io.ReadFull(br, id); err != nil {
				return nil, err
			}
			base, err = p.repo.object(ObjectID(id), false)
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
		if metaOnly {
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

func readOffset(r io.ByteReader) (uint64, error) {
	var offset uint64
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		offset |= uint64(b & 0x7f)
		if b&0x80 == 0 {
			return offset, nil
		}
		offset = (offset + 1) << 7
	}
}

func readBytesAt(r io.ReaderAt, offset int64, len int) []byte {
	buf := make([]byte, len)
	r.ReadAt(buf, offset) // error ignored, unexpected content needs to be handled by caller, just like a corrupted file
	return buf
}

func binarySearch(r io.ReaderAt, tableStart int64, min, max uint32, id ObjectID) (uint32, error) {
	for min <= max {
		mid := min + ((max - min) / 2) // avoid overflow
		midID := ObjectID(readBytesAt(r, tableStart+20*int64(mid), 20))
		if midID == id {
			return mid, nil
		}
		if midID < id {
			min = mid + 1
			continue
		}
		max = mid - 1
	}
	return 0, ObjectNotFound(id)
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