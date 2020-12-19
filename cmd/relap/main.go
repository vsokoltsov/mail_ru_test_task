package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/record"
	"relap/pkg/repositories/storage"
	"relap/pkg/repositories/worker"
	"sync"
)

func main() {
	var (
		path       = flag.String("file", "./../../500.jsonl", "Path to a file")
		resultsDir = flag.String("results", "./../../results/", "Folder with result files.")
		resultExt  = flag.String("ext", "tsv", "Extension of the resulting file.")
		goNum      = flag.Int("go-num", 25, "Number of pages per goroutines")
		wg         = &sync.WaitGroup{}
		jobs       = make(chan models.Job)
		results    = make(chan models.Result)
		errors     = make(chan error)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHandlerHTML()
	recordFile := record.NewRecordFile(
		wg,
		fs,
		jobs,
		results,
		errors,
	)
	workersPool := worker.NewWorkersPool(
		*goNum,
		htmlHandler,
		wg,
		jobs,
		results,
	)

	absPath, filePathErr := filepath.Abs(*path)
	if filePathErr != nil {
		log.Fatalf("Absolute path error: %s", filePathErr)
	}
	file, fileErr := fs.OpenFile(absPath, os.O_RDONLY, 0644)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer file.Close()

	workersPool.StartWorkers()

	readErr := recordFile.ReadLines(file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// TODO: Group data and write to files

	categoryRecords := readFromChannles(results, errors)

	writeWg := &sync.WaitGroup{}
	for category, records := range categoryRecords {
		writeWg.Add(1)
		go func(dir, ext, category string, records []*models.ResultData, wg *sync.WaitGroup, recordFile record.RecordInt) {
			defer wg.Done()
			recordFile.SaveResults(dir, ext, category, records)
			// recordFile.SaveResults(*resultsDir, *resultExt, category, records)
		}(*resultsDir, *resultExt, category, records, writeWg, recordFile)
	}
	writeWg.Done()

	fmt.Println(categoryRecords)
	fmt.Println("Finished")
}

func readFromChannles(results chan models.Result, errors chan error) map[string][]*models.ResultData {
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
				log.Fatalf("Error file reading: %s", errChan.Error())
			}
		}
	}
	return categoryRecords
}
