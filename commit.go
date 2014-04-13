package git

import (
	"container/list"
)

// Commit represents a git commit.
type Commit struct {
	Tree
	Id            sha1 // The id of this commit object
	Author        *Signature
	Committer     *Signature
	CommitMessage string

	parents []sha1 // sha1 strings
}

// Return the commit message. Same as retrieving CommitMessage directly.
func (c *Commit) Message() string {
	return c.CommitMessage
}

// Return parent number n (0-based index)
func (c *Commit) Parent(n int) (*Commit, error) {
	id, err := c.ParentId(n)
	if err != nil {
		return nil, err
	}
	parent, err := c.repo.getCommit(id)
	if err != nil {
		return nil, err
	}
	return parent, nil
}

// Return oid of the parent number n (0-based index). Return nil if no such parent exists.
func (c *Commit) ParentId(n int) (id sha1, err error) {
	if n >= len(c.parents) {
		err = IdNotExist
		return
	}
	return c.parents[n], nil
}

// Return the number of parents of the commit. 0 if this is the
// root commit, otherwise 1,2,...
func (c *Commit) ParentCount() int {
	return len(c.parents)
}

// Return oid of the (root) tree of this commit.
func (c *Commit) TreeId() sha1 {
	return c.Tree.Id
}

func (c *Commit) CommitsBefore() (*list.List, error) {
	return c.repo.getCommitsBefore(c.Id)
}

func (c *Commit) CommitsBeforeUntil(commitId string) (*list.List, error) {
	ec, err := c.repo.GetCommit(commitId)
	if err != nil {
		return nil, err
	}
	return c.repo.CommitsBetween(c, ec)
}

func (c *Commit) CommitsCount() (int, error) {
	return c.repo.commitsCount(c.Id)
}

func (c *Commit) SearchCommits(keyword string) (*list.List, error) {
	return c.repo.searchCommits(c.Id, keyword)
}

func (c *Commit) CommitsByRange(page int) (*list.List, error) {
	return c.repo.commitsByRange(c.Id, page)
}

/*func (c *Commit) GetCommitOfRelPath(relPath string) (*Commit, error) {
	return c.repo.getCommitOfRelPath(c.Id, relPath)
}*/

func (c *Commit) GetCommitOfRelPath(relPath string) (*Commit, error) {
	var key = c.Id.String() + ";" + relPath

	mutex.RLock()
	if v, ok := objCommitCache[key]; ok {
		mutex.RUnlock()
		return c.repo.getCommit(v)
	}
	mutex.RUnlock()

	newc, err := c.repo.getCommitOfRelPath(c.Id, relPath)
	if err != nil {
		return nil, err
	}

	mutex.Lock()
	defer mutex.Unlock()
	objCommitCache[key] = newc.Id
	return newc, nil
}
