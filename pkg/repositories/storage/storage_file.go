package storage

import "os"

type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) OpenFile(path string) (*os.File, error) {
	return os.Open(path)
}

func (fs *FileStorage) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}
