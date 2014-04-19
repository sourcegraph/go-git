package git

import (
	"os"
	"path/filepath"
	"strings"
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
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}
		names = append(names, fi.Name())
	}

	return names, nil
}
