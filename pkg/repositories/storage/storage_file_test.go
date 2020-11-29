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
