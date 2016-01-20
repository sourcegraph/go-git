package git

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	IdNotExist = errors.New("sha1 id not exist")
)

type sha1 [20]byte

// Return string (hex) representation of the Oid
func (s sha1) String() string {
	result := make([]byte, 0, 40)
	hexvalues := []byte("0123456789abcdef")
	for i := 0; i < 20; i++ {
		result = append(result, hexvalues[s[i]>>4])
		result = append(result, hexvalues[s[i]&0xf])
	}
	return string(result)
}

func IsSha1(sha1 string) bool {
	if len(sha1) != 40 {
		return false
	}

	_, err := hex.DecodeString(sha1)
	if err != nil {
		return false
	}

	return true
}

// Create a new sha1 from a Sha1 string of length 40.
func NewIdFromString(s string) (sha1, error) {
	s = strings.TrimSpace(s)
	var id sha1
	if len(s) != 40 {
		return id, fmt.Errorf("Length must be 40")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}

	return NewId(b)
}

// Create a new sha1 from a 20 byte slice.
func NewId(b []byte) (sha1, error) {
	var id sha1
	if len(b) != 20 {
		return id, errors.New("Length must be 20")
	}

	for i := 0; i < 20; i++ {
		id[i] = b[i]
	}
	return id, nil
}
