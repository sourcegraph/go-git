package git

import ()

type Blob struct {
	*TreeEntry

	data   []byte
	dataed bool

	size  int64
	sized bool
}

func (b *Blob) Size() int64 {
	if b.sized {
		return b.size
	}

	size, err := b.ptree.repo.objectSize(b.Id)
	if err != nil {
		return 0
	}

	b.sized = true
	b.size = size
	return b.size
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
