package git

import (
	"io"
)

type Blob struct {
	*TreeEntry
}

func (b *Blob) Data() (io.ReadCloser, error) {
	_, _, dataRc, err := b.ptree.repo.getRawObject(b.Id, false)
	if err != nil {
		return nil, err
	}
	return dataRc, nil
}
