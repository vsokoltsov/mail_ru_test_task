package storage

import (
	"log"
	"os"
	"testing"
)

func TestStorageCreateFile(t *testing.T) {
	storage := NewFileStorage()
	_, err := storage.CreateFile("./test")
	if err != nil {
		t.Errorf("Error of creating the file: %s", err)
	}

	deleteErr := os.Remove("./test")
	if deleteErr != nil {
		log.Fatal(deleteErr)
	}
}

func TestStorageOpenFile(t *testing.T) {
	storage := NewFileStorage()
	_, err := storage.CreateFile("./test")
	if err != nil {
		t.Errorf("Error of creating the file: %s", err)
	}

	_, oErr := storage.OpenFile("./test", os.O_RDONLY, 0644)
	if err != nil {
		t.Errorf("Error of opening fle: %s", oErr)
	}
	deleteErr := os.Remove("./test")
	if deleteErr != nil {
		log.Fatal(deleteErr)
	}
}

func TestStorageResultPath(t *testing.T) {
	storage := NewFileStorage()
	var (
		dir  = "/a/b/c/"
		name = "test"
		ext  = "txt"
	)
	path := storage.ResultPath(dir, name, ext)
	if path != "/a/b/c/test.txt" {
		t.Errorf("Generated file path does not match: %s", path)
	}
}
