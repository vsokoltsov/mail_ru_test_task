package worker

import (
	"os"
	"relap/pkg/models"
)

// WorkersReadPoolInt represents actions for workers read pool interface
type WorkersReadPoolInt interface {
	StartReadWorkers()
	StartWriteWorkers()
	ReadFromChannels(results chan models.Result, writeChan chan PoolWriteJob, errors chan error) (map[string][]*models.ResultData, error)
	listenJobs(id int, jobs <-chan models.Job, results chan<- models.Result)
}

// WorkersWritePoolInt represents actions for workers write pool interface
type WorkersWritePoolInt interface {
	StartWorkers()
	ListenWriteJobs(id int, jobs <-chan models.CategoryJob, results chan<- *os.File)
}

// Int represents actions for worker interface
type Int interface {
	FetchPage(url string, categories []string) (*models.ResultData, error)
}
