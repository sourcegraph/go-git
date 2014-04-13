package git

import (
	"fmt"
	"sort"
)

// There are only a few file modes in Git. They look like unix file modes, but they can only be
// one of these.
const (
	ModeBlob    EntryMode = 0100644
	ModeExec    EntryMode = 0100755
	ModeSymlink EntryMode = 0120000
	ModeCommit  EntryMode = 0160000
	ModeTree    EntryMode = 0040000
)

type EntryMode int

type Entries []*TreeEntry

var sorter = []func(t1, t2 *TreeEntry) bool{
	func(t1, t2 *TreeEntry) bool {
		return t1.IsDir() && !t2.IsDir()
	},
	func(t1, t2 *TreeEntry) bool {
		return t1.Name < t2.Name
	},
}

func (bs Entries) Len() int      { return len(bs) }
func (bs Entries) Swap(i, j int) { bs[i], bs[j] = bs[j], bs[i] }
func (bs Entries) Less(i, j int) bool {
	t1, t2 := bs[i], bs[j]
	var k int
	for k = 0; k < len(sorter)-1; k++ {
		sort := sorter[k]
		switch {
		case sort(t1, t2):
			return true
		case sort(t2, t1):
			return false
		}
	}
	return sorter[k](t1, t2)
}

func (bs Entries) Sort() {
	sort.Sort(bs)
}

type TreeEntry struct {
	Id   sha1
	Name string
	Mode EntryMode
	Type ObjectType

	ptree *Tree

	commit   *Commit
	commited bool
}

func (te *TreeEntry) Blob() *Blob {
	return &Blob{TreeEntry: te}
}

func (te *TreeEntry) IsDir() bool {
	return te.Mode == ModeTree
}

func (te *TreeEntry) IsFile() bool {
	return te.Mode == ModeBlob || te.Mode == ModeExec
}

func (te *TreeEntry) Commit() (*Commit, error) {
	if te.commited {
		return te.commit, nil
	}

	g := te.ptree
	t := te.ptree
	for t != nil {
		g = t
		t = t.ptree
	}

	fmt.Println(g.Id.String())

	return te.commit, nil
}
