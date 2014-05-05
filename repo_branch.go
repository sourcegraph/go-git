package git

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrBranchExisted = errors.New("branch has existed")
)

func IsBranchExist(repoPath, branchName string) bool {
	branchPath := filepath.Join(repoPath, "refs/heads", branchName)
	return isFile(branchPath)
}

func (repo *Repository) IsBranchExist(branchName string) bool {
	return IsBranchExist(repo.Path, branchName)
}

// GetBranches returns all branches of given repository.
func (repo *Repository) GetBranches() ([]string, error) {
	dirPath := filepath.Join(repo.Path, "refs/heads")
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name())
	}

	return names, nil
}

func CreateBranch(repoPath, branchName, id string) error {
	return CreateRef("heads", repoPath, branchName, id)
}

func CreateRef(head, repoPath, branchName, id string) error {
	branchPath := filepath.Join(repoPath, "refs/"+head, branchName)
	if isFile(branchPath) {
		return ErrBranchExisted
	}
	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, id)
	return err
}

func (repo *Repository) CreateBranch(branchName, id string) error {
	return CreateBranch(repo.Path, branchName, id)
}
