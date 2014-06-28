package git

import "io"

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

/*
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
}*/
