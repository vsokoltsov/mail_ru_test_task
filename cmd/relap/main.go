package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/pipeline"
	"relap/pkg/repositories/pool"
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
		readWg       = &sync.WaitGroup{}
		writeWg      = &sync.WaitGroup{}
		readJobs     = make(chan models.ReadJob)
		readResults  = make(chan models.ReadResult)
		errors       = make(chan error)
		writeJobs    = make(chan models.WriteJob)
		writeResults = make(chan models.WriteResult)
	)

	flag.Parse()

	fs := storage.NewFileStorage()
	htmlHandler := handler.NewHTML()
	wrkr := worker.NewWorker(htmlHandler)
	readPool := pool.NewReadPool(*goNum, readWg, readJobs, readResults, wrkr)
	writePool := pool.NewWritePool(*goNum, writeWg, writeJobs, writeResults)

	absPath, filePathErr := filepath.Abs(*path)
	if filePathErr != nil {
		log.Fatalf("Absolute path error: %s", filePathErr)
	}
	file, fileErr := fs.OpenFile(absPath, os.O_RDONLY, 0644)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer file.Close()

	readPool.StartWorkers()
	writePool.StartWorkers()
	readerPipe := pipeline.NewReader(file, readResults, readWg, readJobs, errors)
	writerPipe := pipeline.NewWriter(writeWg, writeJobs, writeResults, errors)
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
