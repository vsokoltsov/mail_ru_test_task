package pool

import (
	"fmt"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/pipeline"
	"relap/pkg/repositories/record"
	"relap/pkg/repositories/worker"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestStartWorkersSuccessResult(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		readWg        = &sync.WaitGroup{}
		readJobs      = make(chan pipeline.ReadJob)
		readResults   = make(chan pipeline.ReadResult)
		workerIntMock = worker.NewMockInt(ctrl)
	)
	workerIntMock.EXPECT().FetchPage("http://example.com", []string{"1", "2", "3"}).Return(&handler.ResultData{
		Title:       "Test",
		Description: "Test",
		URL:         "http://example.com",
		Categories:  []string{"1", "2", "3"},
	}, nil)

	readPool := NewReadPool(1, readWg, readJobs, readResults, workerIntMock)
	readPool.StartWorkers()

	readJobs <- pipeline.ReadJob{Record: &record.Row{URL: "http://example.com", Categories: []string{"1", "2", "3"}}}
	close(readJobs)
	result := <-readResults
	if result.Result.URL != "http://example.com" {
		t.Errorf("Unexpected result of worker. Expected: %s, got: %s", "http://example.com", result.Result.URL)
	}
}

func TestStartWorkersFailedResult(t *testing.T) {
	var (
		ctrl          = gomock.NewController(t)
		readWg        = &sync.WaitGroup{}
		readJobs      = make(chan pipeline.ReadJob)
		readResults   = make(chan pipeline.ReadResult)
		workerIntMock = worker.NewMockInt(ctrl)
	)
	workerIntMock.EXPECT().FetchPage("http://example.com", []string{"1", "2", "3"}).Return(nil, fmt.Errorf("Fetch page error"))

	readPool := NewReadPool(1, readWg, readJobs, readResults, workerIntMock)
	readPool.StartWorkers()

	readJobs <- pipeline.ReadJob{Record: &record.Row{URL: "http://example.com", Categories: []string{"1", "2", "3"}}}
	close(readJobs)
	result := <-readResults
	if result.Err == nil {
		t.Errorf("Expected error, got nil")
	}
	close(readResults)
}
