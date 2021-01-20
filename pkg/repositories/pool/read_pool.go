package pool

import (
	"relap/pkg/models"
	"relap/pkg/repositories/worker"
	"sync"
)

type ReadPool struct {
	workersNum int
	wg         *sync.WaitGroup
	jobs       <-chan models.ReadJob
	results    chan<- models.ReadResult
	worker     worker.Int
}

// ReadJob defines job for read workers pool
type ReadJob struct {
	Record *Record
}

// Result represents operation outcome
type ReadResult struct {
	WorkerID int
	Result   *ResultData
	Err      error
}

func NewReadPool(workersNum int, wg *sync.WaitGroup, jobs <-chan models.ReadJob, results chan<- models.ReadResult, worker worker.Int) Communication {
	return ReadPool{
		wg:         wg,
		workersNum: workersNum,
		jobs:       jobs,
		results:    results,
		worker:     worker,
	}
}

func (rp ReadPool) StartWorkers() {
	rp.wg.Add(rp.workersNum)
	for i := 0; i < rp.workersNum; i++ {
		go func(i int, wp *ReadPool) {
			defer wp.wg.Done()
			wp.listenJobs(i)
		}(i, &rp)
	}
}

func (rp ReadPool) listenJobs(id int) {
	for j := range rp.jobs {
		var (
			result models.ReadResult
		)
		resultData, err := rp.worker.FetchPage(j.Record.URL, j.Record.Categories)
		if err != nil {
			result.Err = err
		}
		result.Result = resultData
		result.WorkerID = id
		rp.results <- result
	}
}
