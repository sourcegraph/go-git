package git

import (
	"path/filepath"
)

func (repo *Repository) IsTagExist(tagName string) bool {
	tagPath := filepath.Join(repo.Path, "refs/tags", tagName)
	return isFile(tagPath)
}

// GetTags returns all tags of given repository.
func (repo *Repository) GetTags() ([]string, error) {
	return repo.readRefDir("refs/tags", "")
}

func (repo *Repository) CreateTag(tagName, idStr string) error {
	return repo.createRef("tags", tagName, idStr)
}

func CreateTag(repoPath, tagName, id string) error {
	return CreateRef("tags", repoPath, tagName, id)
}
