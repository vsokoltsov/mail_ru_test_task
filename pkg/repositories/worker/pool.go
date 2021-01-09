package worker

import (
	"fmt"
	"log"
	"os"
	"relap/pkg/models"
	"strings"
	"sync"
)

// WorkersReadPool represents pool of workers for reading data
type WorkersReadPool struct {
	workersNum    int
	worker        Int
	wg            *sync.WaitGroup
	writeWg       *sync.WaitGroup
	resultWg      *sync.WaitGroup
	jobs          chan models.Job
	results       chan models.Result
	mu            *sync.Mutex
	fmu           *sync.Mutex
	categoryFiles map[string]*os.File
	pwj           chan PoolWriteJob
	writeResults  chan WriteResult
}

// WorkersWritePool represents pool of workers for writing data
type WorkersWritePool struct {
	workersNum int
	jobs       chan models.CategoryJob
	results    chan *os.File
	wg         *sync.WaitGroup
}

type PoolWriteJob struct {
	WorkerID   int
	File       *os.File
	ResultData *models.ResultData
	Category   string
}

type WriteResult struct {
	Category string
	File     *os.File
}

// NewWorkersReadPool returns object that impelements WorkersReadPoolInt
func NewWorkersReadPool(
	workersNum int,
	wg *sync.WaitGroup,
	jobs chan models.Job,
	results chan models.Result,
	worker Int,
	pwj chan PoolWriteJob,
	writeWg *sync.WaitGroup,
	writeResults chan WriteResult,
	resultWg *sync.WaitGroup) WorkersReadPoolInt {
	return WorkersReadPool{
		workersNum:    workersNum,
		wg:            wg,
		writeWg:       writeWg,
		jobs:          jobs,
		results:       results,
		worker:        worker,
		mu:            &sync.Mutex{},
		fmu:           &sync.Mutex{},
		categoryFiles: make(map[string]*os.File),
		pwj:           pwj,
		writeResults:  writeResults,
		resultWg:      resultWg,
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

// StartReadWorkers runs workers for reading
func (wp WorkersReadPool) StartReadWorkers() {
	wp.wg.Add(wp.workersNum)
	for i := 0; i < wp.workersNum; i++ {
		go func(i int, wp *WorkersReadPool) {
			defer wp.wg.Done()
			wp.listenJobs(i, wp.jobs, wp.results)
		}(i, &wp)
	}
}

func (wp WorkersReadPool) StartWriteWorkers() {
	wp.writeWg.Add(wp.workersNum)
	for i := 0; i < wp.workersNum; i++ {
		go func(i int, wp *WorkersReadPool) {
			defer wp.writeWg.Done()
			wp.listenJobsForWrite(i, wp.resultWg, wp.pwj, wp.writeResults)
		}(i, &wp)
	}
}

func (wp WorkersReadPool) listenJobsForWrite(id int, wg *sync.WaitGroup, jobs <-chan PoolWriteJob, results chan<- WriteResult) {
	for j := range jobs {
		j.File.WriteString(strings.Join([]string{j.ResultData.URL, j.ResultData.Title, j.ResultData.Description, "\n"}, " "))
		j.File.Sync()
		results <- WriteResult{Category: j.Category, File: j.File}
	}
}

// ReadFromChannels read data from multiple channels
func (wp WorkersReadPool) ReadFromChannels(results chan models.Result, writeChan chan PoolWriteJob, errors chan error) (map[string][]*models.ResultData, error) {
	wwg := &sync.WaitGroup{}
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
					var (
						catFile *os.File
						fileErr error
					)
					// Write data into intermediate channel
					catFile = wp.getCategoryFile(category)
					if catFile == nil {
						catFile, fileErr = wp.setCategoryFile(category)
						if fileErr != nil {
							return nil, fileErr
						}
					}
					// writeChan <- PoolWriteJob{File: catFile, ResultData: r.Result, Category: category}
					wwg.Add(1)
					go func(wwg *sync.WaitGroup, rwg *sync.WaitGroup, writeChan chan PoolWriteJob, file *os.File, rd *models.ResultData) {
						defer wwg.Done()
						rwg.Add(1)
						writeChan <- PoolWriteJob{File: file, ResultData: rd, Category: category}
					}(wwg, wp.resultWg, writeChan, catFile, r.Result)
					// if ok {

					// } else {
					// 	wp.mu.Lock()
					// 	categoryFile, err := os.Create(category + ".tsv")
					// 	if err != nil {
					// 		wp.mu.Unlock()
					// 		return nil, fmt.Errorf("Error of creating %s file: %s", category, err)
					// 	}
					// 	wp.categoryFiles[category] = categoryFile
					// 	wp.mu.Unlock()
					// }
					// _, ok := categoryRecords[category]
					// if ok {
					// 	categoryRecords[category] = append(categoryRecords[category], r.Result)
					// } else {
					// 	categoryRecords[category] = []*models.ResultData{
					// 		r.Result,
					// 	}
					// }
				}
			}
		case errChan, ok := <-errors:
			if !ok {
				break READ
			}
			if errChan != nil {
				return nil, errChan
			}
		}
	}
	go func(wg *sync.WaitGroup, rwg *sync.WaitGroup, wc chan PoolWriteJob, rc chan WriteResult) {
		wg.Wait()
		rwg.Wait()
		close(wc)
		close(rc)
	}(wwg, wp.resultWg, writeChan, wp.writeResults)
	return categoryRecords, nil
}

func (wp WorkersReadPool) getCategoryFile(category string) *os.File {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.categoryFiles[category]
}

func (wp WorkersReadPool) setCategoryFile(category string) (*os.File, error) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	categoryFile, err := os.Create(category + ".tsv")
	if err != nil {
		return nil, fmt.Errorf("Error of creating %s file: %s", category, err)
	}
	return categoryFile, nil
}
