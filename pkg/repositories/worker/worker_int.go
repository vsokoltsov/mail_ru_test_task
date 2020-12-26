package worker

import (
	"os"
	"relap/pkg/models"
)

type WorkersReadPoolInt interface {
	StartWorkers()
	ReadFromChannels(results chan models.Result, errors chan error) (map[string][]*models.ResultData, error)
	listenJobs(id int, jobs <-chan models.Job, results chan<- models.Result)
}

type WorkersWritePoolInt interface {
	StartWorkers()
	ListenWriteJobs(id int, jobs <-chan models.CategoryJob, results chan<- *os.File)
}

type WorkerInt interface {
	FetchPage(url string, categories []string) (*models.ResultData, error)
}
