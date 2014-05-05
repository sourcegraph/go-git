package git

import (
	"bytes"
	"compress/zlib"
	sha "crypto/sha1"
	"fmt"
	"io"
)

type Blob struct {
	*TreeEntry

	data   []byte
	dataed bool
}

func (b *Blob) Data() ([]byte, error) {
	if b.dataed {
		return b.data, nil
	}
	_, _, data, err := b.ptree.repo.getRawObject(b.Id)
	if err != nil {
		return nil, err
	}
	b.data = data
	b.dataed = true
	return b.data, nil
}

func (b *Blob) Save(w io.Writer, compress bool) (string, error) {
	var data []byte
	buf := bytes.NewBuffer(data)
	header := fmt.Sprintf("blob %d\\0", len(b.data))
	_, err := buf.Write([]byte(header))
	if err != nil {
		return "", err
	}
	_, err = buf.Write(b.data)
	if err != nil {
		return "", err
	}
	if compress {
		var b bytes.Buffer
		c := zlib.NewWriter(&b)
		_, err = io.Copy(c, buf)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(w, &b)
	} else {
		_, err = io.Copy(w, buf)
	}
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha.Sum(data)), nil
}
