package git

import (
	"errors"
	"os"
)

func (repo *Repository) _GetBlob(idStr string) (*Blob, error) {
	id, err := NewIdFromString(idStr)
	if err != nil {
		return nil, err
	}

	return repo.getBlob(id)
}

// Find the blob object in the repository.
func (repo *Repository) getBlob(id sha1) (*Blob, error) {
	var (
		data []byte
		err  error
	)

	// first we need to find out where the commit is stored
	objPath := filepathFromSHA1(repo.Path, id.String())
	_, err = os.Stat(objPath)
	if os.IsNotExist(err) {
		// doesn't exist, let's look if we find the object somewhere else
		for _, indexfile := range repo.indexfiles {
			if offset := indexfile.offsetValues[id]; offset != 0 {
				_, err = os.Stat(indexfile.packpath)
				if os.IsNotExist(err) {
					return nil, err
				}
				_, _, data, err = readObjectBytes(indexfile.packpath, offset, false)
				if err != nil {
					return nil, err
				}
			}
		}
		if len(data) == 0 {
			return nil, errors.New("Object not found")
		}
	} else {
		_, _, data, err = readObjectFile(objPath, false)
		if err != nil {
			return nil, err
		}
	}

	b := new(Blob)
	b.data = data
	return b, nil
}
