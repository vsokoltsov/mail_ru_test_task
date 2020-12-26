package worker

import (
	"bufio"
	"fmt"
	"os"
	"relap/pkg/models"
	"sync"
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestWorkersReadPoolSuccessListenJobs(t *testing.T) {
	var (
		ctrl               = gomock.NewController(t)
		jobs               = make(chan models.Job)
		results            = make(chan models.Result)
		wg                 = &sync.WaitGroup{}
		wrkr               = NewMockInt(ctrl)
		pool               = NewWorkersReadPool(1, wg, jobs, results, wrkr)
		expectedResultData = &models.ResultData{
			Title:       "Title",
			Description: "Description",
			URL:         "http://example.com",
			Categories:  []string{"1", "2", "3"},
		}
	)
	defer ctrl.Finish()
	pool.StartWorkers()

	wrkr.EXPECT().
		FetchPage("http://example.com", []string{"1", "2", "3"}).
		Return(expectedResultData, nil)

	jobs <- models.Job{Record: &models.Record{
		URL:        "http://example.com",
		Categories: []string{"1", "2", "3"},
	}}
	close(jobs)
	res := <-results
	if res.Result.Title != expectedResultData.Title {
		t.Errorf("Mismatched title values: got %s, expected %s", res.Result.Title, expectedResultData.Title)
	}
}

func TestWorkersReadPoolFailedListenJobs(t *testing.T) {
	var (
		ctrl    = gomock.NewController(t)
		jobs    = make(chan models.Job)
		results = make(chan models.Result)
		wg      = &sync.WaitGroup{}
		wrkr    = NewMockInt(ctrl)
		pool    = NewWorkersReadPool(1, wg, jobs, results, wrkr)
	)
	defer ctrl.Finish()
	pool.StartWorkers()

	wrkr.EXPECT().
		FetchPage("http://example.com", []string{"1", "2", "3"}).
		Return(nil, fmt.Errorf("FetchPage error"))

	jobs <- models.Job{Record: &models.Record{
		URL:        "http://example.com",
		Categories: []string{"1", "2", "3"},
	}}
	close(jobs)
	res := <-results
	if res.Err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestWorkerWritePoolSuccessListenJobs(t *testing.T) {
	var (
		ctrl    = gomock.NewController(t)
		jobs    = make(chan models.CategoryJob)
		results = make(chan *os.File)
		wg      = &sync.WaitGroup{}
		pool    = NewWorkersWritePool(1, jobs, results, wg)
		count   int
	)
	defer ctrl.Finish()
	pool.StartWorkers()

	f, _ := os.Create("./test")
	defer os.Remove("./test")
	jobs <- models.CategoryJob{
		File: f,
		ResultsData: []*models.ResultData{
			&models.ResultData{
				Title:       "Title",
				Description: "Description",
				URL:         "http://example.com",
				Categories:  []string{"1", "2", "3"},
			},
		},
	}
	close(jobs)
	res := <-results

	res.Close()
	of, _ := os.Open("./test")
	scanner := bufio.NewScanner(of)
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("Error file reading")
	}

	if count != 1 {
		t.Errorf("Wrong number of lines: expected 1, got: %d", count)
	}

}

func TestWorkersReadPoolSuccessReadFromChannels(t *testing.T) {
	var (
		ctrl    = gomock.NewController(t)
		jobs    = make(chan models.Job)
		results = make(chan models.Result, 3)
		wg      = &sync.WaitGroup{}
		wrkr    = NewMockInt(ctrl)
		pool    = NewWorkersReadPool(1, wg, jobs, results, wrkr)
		err     = make(chan error, 1)
	)
	results <- models.Result{
		Result: &models.ResultData{
			Title:       "Title",
			Description: "Description",
			URL:         "http://example.com",
			Categories:  []string{"1", "2"},
		},
	}
	results <- models.Result{
		Result: &models.ResultData{
			Title:       "Title 2",
			Description: "Description 2",
			URL:         "http://example2.com",
			Categories:  []string{"1"},
		},
	}
	results <- models.Result{
		Err: fmt.Errorf("Error data"),
	}
	close(results)
	close(err)
	categoryMaps, _ := pool.ReadFromChannels(results, err)
	if len(categoryMaps) != 2 {
		t.Errorf("Wrong number of categories, expected - 2, got - %d", len(categoryMaps))
	}

}

func TestWorkersReadPoolFailedReadFromChannels(t *testing.T) {
	var (
		ctrl    = gomock.NewController(t)
		jobs    = make(chan models.Job)
		results = make(chan models.Result, 1)
		wg      = &sync.WaitGroup{}
		wrkr    = NewMockInt(ctrl)
		pool    = NewWorkersReadPool(1, wg, jobs, results, wrkr)
		err     = make(chan error, 1)
	)
	results <- models.Result{
		Err: fmt.Errorf("Error data"),
	}
	err <- fmt.Errorf("Chan error")
	close(results)
	close(err)
	_, fileReadErr := pool.ReadFromChannels(results, err)

	if fileReadErr == nil {
		t.Errorf("Expected error, got nil")
	}
}
