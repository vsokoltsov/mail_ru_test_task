package pipeline

import (
	"os"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/worker"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSuccessCall(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		readWg        = &sync.WaitGroup{}
		readJobs      = make(chan ReadJob)
		readResults   = make(chan ReadResult)
		errors        = make(chan error)
		in            = make(chan interface{})
		out           = make(chan interface{})
		workerIntMock = worker.NewMockInt(ctrl)
	)
	go func(jobs chan ReadJob, results chan ReadResult, wg *sync.WaitGroup) {
		defer wg.Done()
		for range jobs {
			results <- ReadResult{
				Result: &handler.ResultData{
					Title:      "test",
					URL:        "http://example.com",
					Categories: []string{"1", "2", "3"},
				},
			}
		}
	}(readJobs, readResults, readWg)

	workerIntMock.EXPECT().FetchPage("http://example.com", []string{"1", "2", "3"}).Return(&handler.ResultData{
		Title:       "Test",
		Description: "Test",
		URL:         "http://example.com",
		Categories:  []string{"1", "2", "3"},
	}, nil)

	exampleFile, _ := os.Open("test_reader.jsonl")
	defer exampleFile.Close()

	readWg.Add(1)
	reader := NewReader(exampleFile, readResults, readWg, readJobs, errors)
	go reader.Call(in, out)

	data := <-out
	result := data.(*handler.ResultData)
	if result.URL != "http://example.com" {
		t.Errorf("URLS does no match")
	}
}

func TestCallWithError(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		readWg        = &sync.WaitGroup{}
		readJobs      = make(chan ReadJob)
		readResults   = make(chan ReadResult)
		errors        = make(chan error)
		in            = make(chan interface{})
		out           = make(chan interface{})
		workerIntMock = worker.NewMockInt(ctrl)
	)
	go func(jobs chan ReadJob, results chan ReadResult, wg *sync.WaitGroup) {
		defer wg.Done()
		for range jobs {
			results <- ReadResult{
				Result: &handler.ResultData{
					Title:      "test",
					URL:        "http://example.com",
					Categories: []string{"1", "2", "3"},
				},
			}
		}
	}(readJobs, readResults, readWg)

	workerIntMock.EXPECT().FetchPage("http://example.com", []string{"1", "2", "3"}).Return(&handler.ResultData{
		Title:       "Test",
		Description: "Test",
		URL:         "http://example.com",
		Categories:  []string{"1", "2", "3"},
	}, nil)

	exampleFile, _ := os.Open("test_reader.jsonl")
	defer exampleFile.Close()

	// readPool := pool.NewReadPool(1, readWg, readJobs, readResults, workerIntMock)
	// readPool.StartWorkers()
	readWg.Add(1)
	reader := NewReader(exampleFile, readResults, readWg, readJobs, errors)
	go reader.Call(in, out)

	data := <-errors
	if data == nil {
		t.Errorf("Expected error")
	}
}
