package storage

import (
	"os"
	"path/filepath"
)

// FileStorage implements storage Int interface
type FileStorage struct {
	resultExt string
	resultDir string
}

// NewFileStorage returns new instance of FileStorage struct
func NewFileStorage(resultDir, resultExt string) *FileStorage {
	return &FileStorage{
		resultDir: resultDir,
		resultExt: resultExt,
	}
}

// OpenFile opens file with specific directives
func (fs *FileStorage) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// CreateFile creates new file
func (fs *FileStorage) CreateFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

// ResultPath generates file path based on attributes
func (fs *FileStorage) ResultPath(name string) string {
	return filepath.Join(fs.resultDir, name+"."+fs.resultExt)
}
