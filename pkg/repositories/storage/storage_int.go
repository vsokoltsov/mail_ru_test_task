package storage

import "os"

type StorageInt interface {
	OpenFile(path string, flag int, perm os.FileMode) (*os.File, error)
	CreateFile(path string) (*os.File, error)
	ResultPath(dir, name, ext string) string
}
