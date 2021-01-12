package pipeline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"relap/pkg/models"
	"sync"
)

type Reader struct {
	file    *os.File
	results chan models.ReadResult
	wg      *sync.WaitGroup
	jobs    chan models.ReadJob
	errors  chan error
}

func NewReader(file *os.File, results chan models.ReadResult, wg *sync.WaitGroup, jobs chan models.ReadJob, errors chan error) Pipe {
	return Reader{
		file:    file,
		results: results,
		jobs:    jobs,
		errors:  errors,
		wg:      wg,
	}
}

func (r Reader) Call(in, out chan interface{}) {
	mainWg := &sync.WaitGroup{}
	recordWg := &sync.WaitGroup{}
	go func(file *os.File, jobs chan models.ReadJob, errors chan error, wg *sync.WaitGroup, mainWg *sync.WaitGroup) {
		defer close(jobs)
		defer close(errors)

		scanner := bufio.NewScanner(file)
		var writesNum int
		for scanner.Scan() {
			bytes := scanner.Bytes()
			record, decodeError := decodeLine(bytes)
			if decodeError != nil {
				errors <- decodeError
				break
			}

			if len(record.Categories) > 0 {
				writesNum++
				jobs <- models.ReadJob{Record: record}
			}
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			errors <- scannerErr
		}
	}(r.file, r.jobs, r.errors, recordWg, mainWg)

	go func(wg *sync.WaitGroup, results chan models.ReadResult) {
		wg.Wait()
		close(results)
	}(r.wg, r.results)

	for res := range r.results {
		if res.Err == nil {
			out <- res.Result
		}
	}
}

func decodeLine(bytes []byte) (*models.Record, error) {
	var record models.Record
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}
