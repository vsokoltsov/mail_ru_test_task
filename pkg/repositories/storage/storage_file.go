package storage

import "os"

type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(path, flag, perm)
}

func (fs *FileStorage) CreateFile(path string) (*os.File, error) {
	return os.Create(path)
}

func (fs *FileStorage) ResultPath(dir, name, ext string) string {
	return dir + name + "." + ext
}
