package storage

import "os"

type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) OpenFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
}

func (fs *FileStorage) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

func (fs *FileStorage) ResultPath(dir, name, ext string) string {
	return dir + name + "." + ext
}
