package worker

import (
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"sync"
)

type WorkersPool struct {
	workersNum int
	handler    handler.HandlerInt
	worker     WorkerInt
	wg         *sync.WaitGroup
	jobs       chan models.Job
	results    chan models.Result
}

func NewWorkersPool(
	workersNum int,
	handler handler.HandlerInt,
	wg *sync.WaitGroup,
	jobs chan models.Job,
	results chan models.Result) WorkersPoolInt {
	return WorkersPool{
		workersNum: workersNum,
		handler:    handler,
		wg:         wg,
		jobs:       jobs,
		results:    results,
	}
}

func (wp WorkersPool) listenJobs(id int, jobs <-chan models.Job, results chan<- models.Result) {
	for j := range jobs {
		var (
			result models.Result
			worker = NewWorker(id, wp.handler)
		)
		resultData, err := worker.FetchPage(j.Record.URL, j.Record.Categories)
		if err != nil {
			result.Err = err
		}
		result.Result = resultData
		result.WorkerID = id
		results <- result
	}
}

func (wp WorkersPool) StartWorkers() {
	wp.wg.Add(wp.workersNum)
	for i := 0; i < wp.workersNum; i++ {
		go func(i int, wp *WorkersPool) {
			defer wp.wg.Done()
			wp.listenJobs(i, wp.jobs, wp.results)
		}(i, &wp)
	}
}

// func (wp WorkersPool) fetchPage(url string, categories []string) (*models.ResultData, error) {
// 	resp, respErr := wp.client.Get(url)
// 	if respErr != nil {
// 		return nil, respErr
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNotFound {
// 		data := &models.ResultData{
// 			Title:       "Not Found",
// 			Description: "Not Found",
// 			URL:         url,
// 			Categories:  categories,
// 		}
// 		return data, nil
// 	}
// 	resultData, resultErr := wp.handler.Parse(resp.Body)
// 	if resultErr != nil {
// 		return nil, resultErr
// 	}
// 	resultData.URL = url
// 	resultData.Categories = categories
// 	return resultData, nil
// }
