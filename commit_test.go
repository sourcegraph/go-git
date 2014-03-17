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
	"testing"
)

// Guard for runtime error: slice bounds out of range
func TestParseCommitData(t *testing.T) {
	str := "tree 47e960bd3b10e549716c31badb1fc06aacd708e1\n" +
		"author Artiom <kron@example.com> 1379666165 +0300" +
		"committer Artiom <kron@example.com> 1379666165 +0300\n\n" +
		"if case if ClientForAction will return error, client can absent (be nil)\n\n" +
		"Conflicts:\n" +
		"	app/class.js\n"

	commit, _ := parseCommitData([]byte(str))

	if commit.treeId.String() != "47e960bd3b10e549716c31badb1fc06aacd708e1" {
		t.Fatalf("Got bad tree %s", commit.treeId)
	}
}
