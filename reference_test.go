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
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAllReferences(t *testing.T) {
	Convey("Get all repository references", t, func() {
		//repo, err := OpenRepository("_testdata/testrepo.git")
		repo, err := OpenRepository("/Users/lunny/git/gogs-repositories/xiaoxiao/test.git")
		So(err, ShouldBeNil)

		refs, err := repo.AllReferencesMap()

		So(err, ShouldBeNil)
		So(len(refs), ShouldBeGreaterThan, 0)

		So(refs["master"].BranchName(), ShouldEqual, "master")
		So(refs["dev"].BranchName(), ShouldEqual, "dev")

		for name, ref := range refs {
			fmt.Println(name, ref)
		}
	})
}

func TestCurrentReference(t *testing.T) {
	Convey("Get Current repository references", t, func() {
		repo, err := OpenRepository("_testdata/testrepo.git")
		So(err, ShouldBeNil)

		ref, err := repo.CurrentReference()
		So(err, ShouldBeNil)
		So(ref.Name, ShouldEqual, "master")
	})
}
