package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/pipeline"
	"relap/pkg/repositories/pool"
	"relap/pkg/repositories/storage"
	"relap/pkg/repositories/worker"
	"strings"
	"sync"
	"time"
)

func getProjectPath(delimeter string) string {
	projectDirectory, directoryErr := os.Getwd()

	if directoryErr != nil {
		log.Fatalf("Could not locate current directory: %s", directoryErr)
	}

	isUnderCmd := strings.Contains(projectDirectory, delimeter)
	if isUnderCmd {
		var cmdIdx int
		splitPath := strings.Split(projectDirectory, "/")
		for idx, pathElem := range splitPath {
			if pathElem == delimeter {
				cmdIdx = idx
				break
			}
		}
		projectDirectory = strings.Join(splitPath[:cmdIdx], "/")
	}

	return projectDirectory
}

func main() {
	var (
		startTime    = time.Now()
		filePath     = flag.String("file", "500.jsonl", "Path to a file")
		resultsDir   = flag.String("results", "results/", "Folder with result files.")
		resultExt    = flag.String("ext", "tsv", "Extension of the resulting file.")
		goNum        = flag.Int("go-num", 25, "Number of pages per goroutines")
		readWg       = &sync.WaitGroup{}
		writeWg      = &sync.WaitGroup{}
		readJobs     = make(chan pipeline.ReadJob)
		readResults  = make(chan pipeline.ReadResult)
		errors       = make(chan error)
		writeJobs    = make(chan pipeline.WriteJob)
		writeResults = make(chan pipeline.WriteResult)
	)

	flag.Parse()

	// Define base project path
	projectPath := getProjectPath("cmd")
	resultPath := filepath.Join(projectPath, *resultsDir)
	fs := storage.NewFileStorage(resultPath, *resultExt)

	// Initialize HTML handler
	htmlHandler := handler.NewHTML()

	// Initialize worker for fetching pages
	wrkr := worker.NewWorker(htmlHandler)

	// Initializing read and write pool
	readPool := pool.NewReadPool(*goNum, readWg, readJobs, readResults, wrkr)
	writePool := pool.NewWritePool(*goNum, writeWg, writeJobs, writeResults)

	// Opens initial file
	absPath := filepath.Join(projectPath, *filePath)
	file, fileErr := fs.OpenFile(absPath, os.O_RDONLY, 0644)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer file.Close()

	// Run workers pool for read and write operations
	readPool.StartWorkers()
	writePool.StartWorkers()

	// Define pipes for all operations
	readerPipe := pipeline.NewReader(file, readResults, readWg, readJobs, errors)
	writerPipe := pipeline.NewWriter(writeWg, writeJobs, writeResults, errors, fs)
	ReducerPipe := pipeline.NewReducer()

	pipes := []pipeline.Pipe{
		readerPipe,
		writerPipe,
		ReducerPipe,
	}

	// Execute pipeline
	pipeline.ExecutePipeline(pipes...)
	elapsed := time.Since(startTime)
	log.Printf("Program has finished. It took %s", elapsed)
}
