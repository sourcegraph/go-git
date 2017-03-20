package git

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrBranchExisted = errors.New("branch has existed")
)

func IsBranchExist(repoPath, branchName string) bool {
	branchPath := filepath.Join(repoPath, "refs/heads", branchName)
	return isFile(branchPath)
}

func (repo *Repository) IsBranchExist(branchName string) bool {
	branchPath := filepath.Join(repo.Path, "refs/heads", branchName)
	return isFile(branchPath)
}

func (repo *Repository) GetBranches() ([]string, error) {
	return repo.readRefDir("refs/heads", "", true)
}

func (repo *Repository) CreateBranch(branchName, idStr string) error {
	return repo.createRef("heads", branchName, idStr)
}

func (repo *Repository) createRef(head, branchName, idStr string) error {
	branchPath := filepath.Join(repo.Path, "refs/"+head, branchName)
	if isFile(branchPath) {
		return ErrBranchExisted
	}

	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.WriteString(f, idStr)
	return err
}
func (repo *Repository) readRefDir(prefix, relPath string, checkPackedRef bool) ([]string, error) {
	dirPath := filepath.Join(repo.Path, prefix, relPath)
	f, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, fi := range fis {
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}

		relFileName := filepath.Join(relPath, fi.Name())
		if fi.IsDir() {
			subnames, err := repo.readRefDir(prefix, relFileName, false)
			if err != nil {
				return nil, err
			}
			names = append(names, subnames...)
			continue
		}

		names = append(names, relFileName)
	}

	if checkPackedRef {
		unpacked := make(map[string]bool)
		for _, name := range names {
			unpacked[name] = true
		}
		packed, err := repo.readPackedRefs(prefix, relPath)
		if err != nil {
			return names, nil
		}
		for _, packedName := range packed {
			if !unpacked[packedName] {
				names = append(names, packedName)
				unpacked[packedName] = true
			}
		}
	}

	return names, nil
}

func (repo *Repository) readPackedRefs(prefix, relPath string) ([]string, error) {
	path := filepath.Join(repo.Path, "packed-refs")
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return nil, ErrNoPackedRefs
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	refs := make([]string, 0)
	refpath := filepath.Join(prefix, relPath)
	for scan.Scan() {
		if line := scan.Text(); strings.Contains(line, refpath) {
			relFileName, err := filepath.Rel(prefix, line[41:])
			if err != nil {
				continue
			}
			refs = append(refs, relFileName)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return refs, nil
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
