// Copyright (c) 2013 Patrick Gundlach, speedata (Berlin, Germany)
// Copyright 2014 The Gogs Authors.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package git

import (
	"errors"
	"os"
	"time"
)

type Blob struct {
	data    []byte
	modTime time.Time
}

func (b *Blob) ModTime() time.Time {
	return b.modTime
}

// Find the blob object in the repository.
func (repos *Repository) LookupBlob(oid *Oid) (*Blob, error) {
	var (
		data  []byte
		err   error
		fInfo os.FileInfo
	)

	// first we need to find out where the commit is stored
	objPath := filepathFromSHA1(repos.Path, oid.String())
	fInfo, err = os.Stat(objPath)
	if os.IsNotExist(err) {
		// doesn't exist, let's look if we find the object somewhere else
		for _, indexfile := range repos.indexfiles {
			if offset := indexfile.offsetValues[oid.Bytes]; offset != 0 {
				fInfo, err = os.Stat(indexfile.packpath)
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
	b.modTime = fInfo.ModTime()

	return b, nil
}

func (b *Blob) Size() int {
	return len(b.data)
}

func (b *Blob) Contents() []byte {
	return b.data
}
