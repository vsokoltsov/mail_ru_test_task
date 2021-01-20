package pool

import (
	"relap/pkg/repositories/pipeline"
	"relap/pkg/repositories/worker"
	"sync"
)

type ReadPool struct {
	workersNum int
	wg         *sync.WaitGroup
	jobs       <-chan pipeline.ReadJob
	results    chan<- pipeline.ReadResult
	worker     worker.Int
}

func NewReadPool(workersNum int, wg *sync.WaitGroup, jobs <-chan pipeline.ReadJob, results chan<- pipeline.ReadResult, worker worker.Int) Communication {
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
			result pipeline.ReadResult
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
