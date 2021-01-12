package storage

import (
	"os"
	"path/filepath"
)

// FileStorage implements storage Int interface
type FileStorage struct{}

// NewFileStorage returns new instance of FileStorage struct
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// OpenFile opens file with specific directives
func (fs *FileStorage) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// CreateFile creates new file
func (fs *FileStorage) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

// ResultPath generates file path based on attributes
func (fs *FileStorage) ResultPath(dir, name, ext string) string {
	return filepath.Join(dir, name+"."+ext)
}
