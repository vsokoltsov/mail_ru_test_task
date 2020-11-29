package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"sync"
)

var counterCalls int

type RecordFile struct {
	handler     handler.HandlerInt
	wg          *sync.WaitGroup
	resultsChan chan *models.ResultData
	errorsChan  chan error
	goNum       int
}

func NewRecordFile(
	handler handler.HandlerInt,
	wg *sync.WaitGroup,
	resultsChan chan *models.ResultData,
	errorsChan chan error,
	goNum int) *RecordFile {
	return &RecordFile{
		handler:     handler,
		resultsChan: resultsChan,
		errorsChan:  errorsChan,
		goNum:       goNum,
		wg:          wg,
	}
}

func (rf *RecordFile) ReadLines(file *os.File) error {
	var (
		counter int
		lines   int
		records []*models.Record
	)

	scanner := bufio.NewScanner(file)
	wg := &sync.WaitGroup{}
	for scanner.Scan() {
		bytes := scanner.Bytes()
		record, decodeError := rf.decodeLine(bytes)
		if decodeError != nil {
			return decodeError
		}

		lines++
		counter++
		records = append(records, record)
		if counter == rf.goNum {
			wg.Add(1)
			go rf.fetchPages(records)
			counter = 0
			records = []*models.Record{}
		}
	}

	if scannerErr := scanner.Err(); scannerErr != nil {
		return scannerErr
	}

	go func(wg *sync.WaitGroup, resultDataChan chan *models.ResultData, errorsChan chan error) {
		wg.Wait()
		fmt.Println("Succesfully waited")
		close(errorsChan)
		close(resultDataChan)
	}(rf.wg, rf.resultsChan, rf.errorsChan)

	log.Printf("Overall lines: %d", lines)
	log.Printf("Go num: %d", rf.goNum)
	return nil
}
func (rf *RecordFile) decodeLine(bytes []byte) (*models.Record, error) {
	var record models.Record
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}

func (rf *RecordFile) fetchPages(records []*models.Record) {
	defer rf.wg.Done()

	for _, rec := range records {
		data, dataErr := rf.fetchPage(rec.URL, rec.Categories)
		if dataErr != nil {
			rf.errorsChan <- dataErr
		} else {
			rf.resultsChan <- data
		}
	}

	fmt.Println("Page has finished")
}

func (rf *RecordFile) fetchPage(url string, categories []string) (*models.ResultData, error) {
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
