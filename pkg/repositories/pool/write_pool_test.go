package pool

import (
	"os"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/pipeline"
	"sync"
	"testing"
)

func TestStartWriteWorkers(t *testing.T) {
	var (
		writeWg     = &sync.WaitGroup{}
		writeJobs   = make(chan pipeline.WriteJob)
		writeResult = make(chan pipeline.WriteResult)
	)
	exampleFile, _ := os.Create("test.txt")
	defer exampleFile.Close()
	defer os.Remove("test.txt")

	writePool := NewWritePool(1, writeWg, writeJobs, writeResult)
	writePool.StartWorkers()

	writeJobs <- pipeline.WriteJob{
		WorkerID: 1,
		File:     exampleFile,
		ResultData: &handler.ResultData{
			URL:         "test",
			Title:       "test",
			Description: "test",
		},
		Category: "test",
	}
	close(writeJobs)
	res := <-writeResult
	if res.Category != "test" {
		t.Errorf("Category does not match")
	}
}
