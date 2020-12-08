package worker

import "relap/pkg/models"

type WorkersPoolInt interface {
	StartWorkers()
	listenJobs(id int, jobs <-chan models.Job, results chan<- models.Result)
}

type WorkerInt interface {
	FetchPage(url string, categories []string) (*models.ResultData, error)
}
