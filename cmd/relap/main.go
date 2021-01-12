package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/pipeline"
	"relap/pkg/repositories/storage"
	"relap/pkg/repositories/worker"
	"sync"
	"time"
)

func main() {
	var (
		startTime = time.Now()
		path      = flag.String("file", "./../../500.jsonl", "Path to a file")
		// resultsDir = flag.String("results", "./../../results/", "Folder with result files.")
		// resultExt  = flag.String("ext", "tsv", "Extension of the resulting file.")
		goNum        = flag.Int("go-num", 25, "Number of pages per goroutines")
		wg           = &sync.WaitGroup{}
		writeWg      = &sync.WaitGroup{}
		jobs         = make(chan models.Job)
		results      = make(chan models.Result)
		errors       = make(chan error)
		pwj          = make(chan worker.PoolWriteJob)
		writeResults = make(chan worker.WriteResult)
		resultWg     = &sync.WaitGroup{}

		// fileJobs = make(chan models.CategoryJob)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHTML()
	wrkr := worker.NewWorker(htmlHandler)
	workersPool := worker.NewWorkersReadPool(
		*goNum,
		wg,
		jobs,
		results,
		wrkr,
		pwj,
		writeWg,
		writeResults,
		resultWg,
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

	workersPool.StartReadWorkers()
	workersPool.StartWriteWorkers()
	readerPipe := pipeline.NewReader(file, results, wg, jobs, errors)
	writerPipe := pipeline.NewWriter(writeWg, pwj, writeResults, errors)
	combinerPipe := pipeline.NewCombiner()

	pipes := []pipeline.Pipe{
		readerPipe,
		writerPipe,
		combinerPipe,
	}
	pipeline.ExecutePipeline(pipes...)
	elapsed := time.Since(startTime)
	log.Printf("Program has finished. It took %s", elapsed)
}
