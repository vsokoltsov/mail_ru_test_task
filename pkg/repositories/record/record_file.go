package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/storage"
	"strings"
	"sync"
	"time"
)

var counterCalls int

type RecordFile struct {
	handler     handler.HandlerInt
	wg          *sync.WaitGroup
	resultsChan chan *models.ResultData
	errorsChan  chan error
	goNum       int
	storage     storage.StorageInt
	client      *http.Client
	jobs        chan models.Job
	results     chan models.Result
}

func NewRecordFile(
	handler handler.HandlerInt,
	wg *sync.WaitGroup,
	resultsChan chan *models.ResultData,
	errorsChan chan error,
	goNum int,
	storage storage.StorageInt,
	jobs chan models.Job,
	results chan models.Result) RecordInt {
	return RecordFile{
		handler:     handler,
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		goNum:       goNum,
		wg:          wg,
		storage:     storage,
		jobs:        jobs,
		results:     results,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (rf RecordFile) worker(id int, jobs <-chan models.Job, results chan<- models.Result, wg *sync.WaitGroup) {
	for j := range jobs {
		var result models.Result
		resultData, err := rf.fetchPage(j.Record.URL, j.Record.Categories)
		if err != nil {
			result.Err = err
		}
		result.Result = resultData
		result.WorkerID = id
		results <- result
	}
}

func (rf RecordFile) ReadLines(file *os.File) error {
	var (
		counter int
	)

	recordWg := &sync.WaitGroup{}
	go func(file *os.File, counter int, jobs chan models.Job, wg *sync.WaitGroup) {
		// wg.Add(1)
		// defer wg.Done()
		scanner := bufio.NewScanner(file)
		var writesNum int
		for scanner.Scan() {
			bytes := scanner.Bytes()
			record, decodeError := rf.decodeLine(bytes)
			if decodeError != nil {
				// return decodeError
			}

			counter++
			if len(record.Categories) > 0 {
				writesNum++
				jobs <- models.Job{Record: record}
				fmt.Println("Number of writes: ", writesNum)
			}
		}

		close(jobs)
		if scannerErr := scanner.Err(); scannerErr != nil {
			// return scannerErr
		}
		// close(jobs)
	}(file, counter, rf.jobs, recordWg)
	// recordWg.Wait()

	go func(wg *sync.WaitGroup, results chan models.Result) {
		wg.Wait()
		close(results)
	}(rf.wg, rf.results)

	return nil
}
func (rf RecordFile) decodeLine(bytes []byte) (*models.Record, error) {
	var record models.Record
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}

func (rf RecordFile) fetchPages(records []*models.Record) {
	defer rf.wg.Done()

	for _, rec := range records {
		data, dataErr := rf.fetchPage(rec.URL, rec.Categories)
		if dataErr != nil {
			rf.errorsChan <- dataErr
		} else {
			rf.resultsChan <- data
		}
	}
}

func (rf RecordFile) fetchPage(url string, categories []string) (*models.ResultData, error) {
	resp, respErr := rf.client.Get(url)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		data := &models.ResultData{
			Title:       "Not Found",
			Description: "Not Found",
			URL:         url,
			Categories:  categories,
		}
		return data, nil
	}
	resultData, resultErr := rf.handler.Parse(resp.Body)
	if resultErr != nil {
		return nil, resultErr
	}
	resultData.URL = url
	resultData.Categories = categories
	return resultData, nil
}

func (rf RecordFile) formCategoriesData(exitChan chan bool) map[string][]*models.ResultData {
	results := make(map[string][]*models.ResultData)
	for {
		select {
		case result := <-rf.resultsChan:
			if result != nil {
				for _, category := range result.Categories {
					_, ok := results[category]
					if ok {
						results[category] = append(results[category], result)
					} else {
						results[category] = []*models.ResultData{result}
					}
				}
			}
		case err := <-rf.errorsChan:
			if err != nil {
				fmt.Printf("Request error: %s", err)
			}
		case <-exitChan:
			return results
		}
	}
}

func (rf RecordFile) SaveResults(dir, ext, categoryName string, results []*models.ResultData) (string, error) {
	var (
		f       *os.File
		fileErr error
	)

	path := rf.storage.ResultPath(dir, categoryName, ext)
	if _, err := os.Stat(path); err == nil {
		f, fileErr = rf.storage.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		f, fileErr = rf.storage.CreateFile(path)
	}
	defer f.Close()

	if fileErr != nil {
		return "", fileErr
	}

	for _, fd := range results {
		f.Write([]byte(strings.Join([]string{fd.URL, fd.Title, fd.Description, "\n"}, " ")))
	}

	f.Sync()

	return path, nil
}
