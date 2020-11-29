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
)

var counterCalls int

type RecordFile struct {
	handler     handler.HandlerInt
	wg          *sync.WaitGroup
	resultsChan chan *models.ResultData
	errorsChan  chan error
	goNum       int
	storage     storage.StorageInt
}

func NewRecordFile(
	handler handler.HandlerInt,
	wg *sync.WaitGroup,
	resultsChan chan *models.ResultData,
	errorsChan chan error,
	goNum int,
	storage storage.StorageInt) RecordInt {
	return RecordFile{
		handler:     handler,
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		goNum:       goNum,
		wg:          wg,
		storage:     storage,
	}
}

func (rf RecordFile) ReadLines(file *os.File) (map[string][]*models.ResultData, error) {
	var (
		counter  int
		lines    int
		records  []*models.Record
		exitChan = make(chan bool)
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		record, decodeError := rf.decodeLine(bytes)
		if decodeError != nil {
			return nil, decodeError
		}

		lines++
		counter++
		records = append(records, record)
		if counter == rf.goNum {
			rf.wg.Add(1)
			go rf.fetchPages(records)
			counter = 0
			records = []*models.Record{}
		}
	}

	if scannerErr := scanner.Err(); scannerErr != nil {
		return nil, scannerErr
	}

	go func(wg *sync.WaitGroup, resultDataChan chan *models.ResultData, errorsChan chan error, exitChan chan bool) {
		wg.Wait()
		close(resultDataChan)
		close(errorsChan)
		exitChan <- true
	}(rf.wg, rf.resultsChan, rf.errorsChan, exitChan)

	results := rf.formCategoriesData(exitChan)
	return results, nil
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
	resp, respErr := http.Get(url)
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
		f, fileErr = rf.storage.OpenFile(path)
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
