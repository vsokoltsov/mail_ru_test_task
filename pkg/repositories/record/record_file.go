package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"relap/pkg/models"
	"relap/pkg/repositories/storage"
	"sync"
)

var counterCalls int

// File implements record.Int
type File struct {
	wg      *sync.WaitGroup
	storage storage.Int
	jobs    chan models.Job
	results chan models.Result
	errors  chan error
}

// NewFile returns new instance of File struct
func NewFile(
	wg *sync.WaitGroup,
	storage storage.Int,
	jobs chan models.Job,
	results chan models.Result,
	errors chan error) Int {
	return File{
		wg:      wg,
		storage: storage,
		jobs:    jobs,
		results: results,
		errors:  errors,
	}
}

// Readlines reads file and write results to channels
func (rf File) ReadLines(file *os.File) error {
	recordWg := &sync.WaitGroup{}
	go func(file *os.File, jobs chan models.Job, errors chan error, wg *sync.WaitGroup) {
		defer close(jobs)
		defer close(errors)

		scanner := bufio.NewScanner(file)
		var writesNum int
		for scanner.Scan() {
			bytes := scanner.Bytes()
			record, decodeError := rf.decodeLine(bytes)
			if decodeError != nil {
				errors <- decodeError
				break
			}

			if len(record.Categories) > 0 {
				writesNum++
				jobs <- models.Job{Record: record}
			}
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			errors <- scannerErr
		}
	}(file, rf.jobs, rf.errors, recordWg)

	go func(wg *sync.WaitGroup, results chan models.Result) {
		wg.Wait()
		close(results)
	}(rf.wg, rf.results)

	return nil
}

// decodeLine decode json line from file
func (rf File) decodeLine(bytes []byte) (*models.Record, error) {
	var record models.Record
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}
