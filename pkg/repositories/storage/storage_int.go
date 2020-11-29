package storage

import "os"

type StorageInt interface {
	OpenFile(path string) (*os.File, error)
	CreateFile(path string) (*os.File, error)
	ResultPath(dir, name, ext string) string
}
