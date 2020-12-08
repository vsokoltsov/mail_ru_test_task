package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/storage"
	"strings"
	"sync"
)

var counterCalls int

type RecordFile struct {
	wg      *sync.WaitGroup
	storage storage.StorageInt
	jobs    chan models.Job
	results chan models.Result
}

func NewRecordFile(
	wg *sync.WaitGroup,
	storage storage.StorageInt,
	jobs chan models.Job,
	results chan models.Result) RecordInt {
	return RecordFile{
		wg:      wg,
		storage: storage,
		jobs:    jobs,
		results: results,
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
