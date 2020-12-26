package worker

import (
	"log"
	"os"
	"relap/pkg/models"
	"strings"
	"sync"
)

// WorkersReadPool represents pool of workers for reading data
type WorkersReadPool struct {
	workersNum int
	worker     Int
	wg         *sync.WaitGroup
	jobs       chan models.Job
	results    chan models.Result
}

// WorkersWritePool represents pool of workers for writing data
type WorkersWritePool struct {
	workersNum int
	jobs       chan models.CategoryJob
	results    chan *os.File
	wg         *sync.WaitGroup
}

// NewWorkersWritePool returns object that impelements WorkersWritePoolInt
func NewWorkersWritePool(workersNum int, jobs chan models.CategoryJob, results chan *os.File, wg *sync.WaitGroup) WorkersWritePoolInt {
	return WorkersWritePool{
		workersNum: workersNum,
		jobs:       jobs,
		results:    results,
		wg:         wg,
	}
}

// NewWorkersReadPool returns object that impelements WorkersReadPoolInt
func NewWorkersReadPool(
	workersNum int,
	wg *sync.WaitGroup,
	jobs chan models.Job,
	results chan models.Result,
	worker Int) WorkersReadPoolInt {
	return WorkersReadPool{
		workersNum: workersNum,
		wg:         wg,
		jobs:       jobs,
		results:    results,
		worker:     worker,
	}
}

// listenJobs Listen jobs from channel and fetch pages
func (wp WorkersReadPool) listenJobs(id int, jobs <-chan models.Job, results chan<- models.Result) {
	for j := range jobs {
		var (
			result models.Result
		)
		resultData, err := wp.worker.FetchPage(j.Record.URL, j.Record.Categories)
		if err != nil {
			result.Err = err
		}
		result.Result = resultData
		result.WorkerID = id
		results <- result
	}
}

// StartWorkers runs workers
func (wp WorkersReadPool) StartWorkers() {
	wp.wg.Add(wp.workersNum)
	for i := 0; i < wp.workersNum; i++ {
		go func(i int, wp *WorkersReadPool) {
			defer wp.wg.Done()
			wp.listenJobs(i, wp.jobs, wp.results)
		}(i, &wp)
	}
}

// StartWorkers runs workers
func (wwp WorkersWritePool) StartWorkers() {
	wwp.wg.Add(wwp.workersNum)
	for i := 0; i < wwp.workersNum; i++ {
		go func(i int, wwp WorkersWritePool) {
			defer wwp.wg.Done()
			wwp.ListenWriteJobs(i, wwp.jobs, wwp.results)
		}(i, wwp)
	}
}

// ListenWriteJobs listen write jobs and write to particular file
func (wwp WorkersWritePool) ListenWriteJobs(id int, jobs <-chan models.CategoryJob, results chan<- *os.File) {
	for j := range jobs {
		for _, fd := range j.ResultsData {
			j.File.WriteString(strings.Join([]string{fd.URL, fd.Title, fd.Description, "\n"}, " "))
		}
		results <- j.File
	}
}

// ReadFromChannels read data from multiple channels
func (wp WorkersReadPool) ReadFromChannels(results chan models.Result, errors chan error) (map[string][]*models.ResultData, error) {
	categoryRecords := make(map[string][]*models.ResultData)
READ:
	for {
		select {
		case r, ok := <-results:
			if !ok {
				break READ
			}
			if r.Err != nil {
				log.Printf("Error: %s", r.Err)
			} else {
				log.Printf("URL: %s; Title: %s; Description: %s", r.Result.URL, r.Result.Title, r.Result.Description)
				for _, category := range r.Result.Categories {
					_, ok := categoryRecords[category]
					if ok {
						categoryRecords[category] = append(categoryRecords[category], r.Result)
					} else {
						categoryRecords[category] = []*models.ResultData{
							r.Result,
						}
					}
				}
			}
		case errChan := <-errors:
			if errChan != nil {
				return nil, errChan
			}
		}
	}
	return categoryRecords, nil
}
