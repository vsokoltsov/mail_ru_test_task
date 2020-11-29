package main

import (
	"flag"
	"fmt"
	"log"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/record"
	"relap/pkg/repositories/storage"
	"sync"
)

func main() {
	var (
		path = flag.String("file", "./../../500.jsonl", "Path to a file")
		// resultsDirPtr  = flag.String("results", "./../../results/", "Folder with result files")
		goNum          = flag.Int("go-num", 25, "Number of pages per goroutines")
		wg             = &sync.WaitGroup{}
		resultDataChan = make(chan *models.ResultData)
		errorsChan     = make(chan error)
		results        = make(map[string][]*models.ResultData)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHandlerHTML()
	recordFile := record.NewRecordFile(
		htmlHandler,
		wg,
		resultDataChan,
		errorsChan,
		*goNum,
	)

	file, fileErr := fs.OpenFile(*path)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer file.Close()

	readErr := recordFile.ReadLines(file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	for chanErr := range errorsChan {
		fmt.Printf("Request error: %s", chanErr)
	}

	for fileData := range resultDataChan {
		if fileData != nil {
			for _, category := range fileData.Categories {
				_, ok := results[category]
				if ok {
					results[category] = append(results[category], fileData)
				} else {
					results[category] = []*models.ResultData{fileData}
				}
			}
		}
	}

	fmt.Println(results)

	fmt.Println("Finished execution")
}
