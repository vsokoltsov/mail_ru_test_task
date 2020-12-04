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
		path = flag.String("file", "./../../500.jsonl", "Path to a file")
		// resultsDir     = flag.String("results", "./../../results/", "Folder with result files.")
		// resultExt      = flag.String("ext", "tsv", "Extension of the resulting file.")
		goNum = flag.Int("go-num", 25, "Number of pages per goroutines")
		wg    = &sync.WaitGroup{}
		// results        []*models.ResultData
		resultDataChan = make(chan *models.ResultData)
		errorsChan     = make(chan error, 1)
		// results        = make(map[string][]*models.ResultData)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHandlerHTML()
	collector := worker.StartDispatcher(*goNum, htmlHandler, resultDataChan, errorsChan)
	recordFile := record.NewRecordFile(
		&collector,
		htmlHandler,
		wg,
		resultDataChan,
		errorsChan,
		*goNum,
		fs,
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

	readErr := recordFile.ReadLines(file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	go func(results chan *models.ResultData, errors chan error) {
		close(results)
		close(errors)
	}(resultDataChan, errorsChan)

	go func(results chan *models.ResultData, errors chan error) {
		for {
			select {
			case res := <-resultDataChan:
				if res != nil {
					fmt.Println(res)
				}
			case err := <-errorsChan:
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}(resultDataChan, errorsChan)
	// collector.End <- true
	// L:
	// for {
	// 	var (
	// 		resEmpty bool
	// 		errEmpty bool
	// 	)
	// 	select {
	// 	case res := <-resultDataChan:
	// 		if res != nil {
	// 			results = append(results, res)
	// 			fmt.Println(res)
	// 		} else {
	// 			resEmpty = true
	// 		}
	// 	case err := <-errorsChan:
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		} else {
	// 			errEmpty = true
	// 		}
	// 	}

	// 	if resEmpty && errEmpty {
	// 		break L
	// 	}
	// }
	fmt.Println("Finished")
	// for category, results := range results {
	// 	path, saveErr := recordFile.SaveResults(*resultsDir, *resultExt, category, results)
	// 	if saveErr != nil {
	// 		log.Fatal(saveErr)
	// 	}
	// 	log.Printf("Path for %s category: %s", category, path)
	// }
}
