package pipeline

import (
	"fmt"
	"os"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/storage"
	"sync"
	"testing"
)

func TestWriterCall(t *testing.T) {
	var (
		in          = make(chan interface{}, 1)
		out         = make(chan interface{})
		writeWg     = &sync.WaitGroup{}
		writeJobs   = make(chan WriteJob)
		writeResult = make(chan WriteResult)
		errors      = make(chan error)
		storageInt  = storage.NewFileStorage("./", "tsv")
	)
	exampleFile, _ := os.OpenFile("test.tsv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	defer exampleFile.Close()
	defer os.Remove("test.tsv")
	defer os.Remove("1.tsv")

	go func(jobs chan WriteJob, results chan WriteResult, wg *sync.WaitGroup, file *os.File) {
		defer wg.Done()
		for range jobs {
			results <- WriteResult{
				Category: "test",
				File:     file,
			}
		}
	}(writeJobs, writeResult, writeWg, exampleFile)

	writeWg.Add(1)
	writer := NewWriter(writeWg, writeJobs, writeResult, errors, storageInt)
	in <- &handler.ResultData{
		Categories:  []string{"1"},
		Title:       "test",
		Description: "test",
		URL:         "http://example.com",
	}
	go writer.Call(in, out)

	res := <-out
	fmt.Println(res)
}
