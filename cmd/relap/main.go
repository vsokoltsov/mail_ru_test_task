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
		goNum   = flag.Int("go-num", 25, "Number of pages per goroutines")
		wg      = &sync.WaitGroup{}
		jobs    = make(chan models.Job)
		results = make(chan models.Result)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHandlerHTML()
	recordFile := record.NewRecordFile(
		wg,
		fs,
		jobs,
		results,
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
	reqNum := 0
	for r := range results {
		reqNum++
		if r.Err != nil {
			fmt.Println("Results loop error: ", r.Err)
		} else {
			fmt.Println("Results loop worker ID:", r.WorkerID, "Result: ", r.Result.URL)
		}
		fmt.Println(reqNum)
	}

	fmt.Println("Finished")
}
