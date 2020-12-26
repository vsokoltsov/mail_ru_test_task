package record

import (
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/storage"
	"sync"
	"testing"
)

func TestSuccessReadLines(t *testing.T) {
	var (
		wg      = &sync.WaitGroup{}
		fs      = storage.NewFileStorage()
		jobs    = make(chan models.Job)
		results = make(chan models.Result)
		errors  = make(chan error)
	)

	file, _ := os.Open("./test.jsonl")
	File := NewFile(
		wg,
		fs,
		jobs,
		results,
		errors,
	)
	readErr := File.ReadLines(file)
	if readErr != nil {
		t.Errorf("ReadLines error: %s", readErr)
	}
}

func TestFailedReadLines(t *testing.T) {
	var (
		wg      = &sync.WaitGroup{}
		fs      = storage.NewFileStorage()
		jobs    = make(chan models.Job)
		results = make(chan models.Result)
		errors  = make(chan error)
	)

	file, _ := os.Open("./test_decode_error.jsonl")
	File := NewFile(
		wg,
		fs,
		jobs,
		results,
		errors,
	)
	readErr := File.ReadLines(file)
	if readErr != nil {
		t.Errorf("ReadLines error: %s", readErr)
	}
	chanErr := <-errors
	if chanErr == nil {
		t.Errorf("Expected error, got nil.")
	}
}
