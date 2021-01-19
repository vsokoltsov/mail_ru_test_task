package pool

import (
	"relap/pkg/models"
	"strings"
	"sync"
)

type WritePool struct {
	workersNum int
	mu         *sync.Mutex
	wg         *sync.WaitGroup
	jobs       <-chan models.WriteJob
	results    chan<- models.WriteResult
}

func NewWritePool(workersNum int, wg *sync.WaitGroup, jobs <-chan models.WriteJob, results chan<- models.WriteResult) Communication {
	return WritePool{
		workersNum: workersNum,
		wg:         wg,
		jobs:       jobs,
		results:    results,
		mu:         &sync.Mutex{},
	}
}

func (wp WritePool) StartWorkers() {
	wp.wg.Add(wp.workersNum)
	for i := 0; i < wp.workersNum; i++ {
		go func(i int, wp *WritePool) {
			defer wp.wg.Done()
			wp.listenJobs(i)
		}(i, &wp)
	}
}

func (wp WritePool) listenJobs(i int) {
	for j := range wp.jobs {
		wp.mu.Lock()
		j.File.Sync()
		j.File.WriteString(strings.Join([]string{j.ResultData.URL, j.ResultData.Title, j.ResultData.Description, "\n"}, " "))
		wp.results <- models.WriteResult{Category: j.Category, File: j.File}
		j.File.Sync()
		wp.mu.Unlock()
	}
}
