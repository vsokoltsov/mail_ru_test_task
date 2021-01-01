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
	File := record.NewFile(
		wg,
		fs,
		jobs,
		results,
		errors,
	)
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

	readErr := File.ReadLines(file)
	if readErr != nil {
		log.Fatal(readErr)
	}
	// workersPool.ReadFromChannels(results, pwj, errors)
	_, readFileError := workersPool.ReadFromChannels(results, pwj, errors)
	if readFileError != nil {
		log.Fatalf("Error file reading: %s", readFileError.Error())
	}
	// go func(wg *sync.WaitGroup, pwj chan worker.PoolWriteJob) {
	// 	wg.Done()
	// 	// close(pwj)
	// }(writeWg, pwj)

	var categoriesFiles = make(map[string]*os.File)
WRITE:
	for {
		select {
		case wr, ok := <-writeResults:
			if !ok {
				break WRITE
			}
			fmt.Println(wr, ok)
			_, exists := categoriesFiles[wr.Category]
			if !exists {
				categoriesFiles[wr.Category] = wr.File
			}
		}
	}
	for _, f := range categoriesFiles {
		f.Close()
	}
	fmt.Println("Finished")
	// categoryFiles := make(map[string]*os.File)
	// for category := range categoryRecords {
	// 	p := filepath.Join(*resultsDir, category+"."+*resultExt)
	// 	categoryAbsPath, categoryFilePathErr := filepath.Abs(p)
	// 	if categoryFilePathErr != nil {
	// 		log.Fatalf("Absolute path error: %s", filePathErr)
	// 	}
	// 	categoryFile, categoryFileErr := fs.CreateFile(categoryAbsPath)
	// 	if categoryFileErr != nil {
	// 		log.Fatalf("Error creating file for category: %s", categoryFileErr)
	// 	}
	// 	_, ok := categoryFiles[category]
	// 	if !ok {
	// 		categoryFiles[category] = categoryFile
	// 	}
	// }

	// fileResults := make(chan *os.File, len(categoryFiles))
	// // writeWg := &sync.WaitGroup{}
	// writePool := worker.NewWorkersWritePool(
	// 	len(categoryRecords),
	// 	fileJobs,
	// 	fileResults,
	// 	writeWg,
	// )
	// writePool.StartWorkers()

	// go func() {
	// 	for category, f := range categoryFiles {
	// 		records := categoryRecords[category]
	// 		fileJobs <- models.CategoryJob{
	// 			File:        f,
	// 			Category:    category,
	// 			ResultsData: records,
	// 		}
	// 	}
	// 	close(fileJobs)
	// }()

	// go func(wg *sync.WaitGroup, writeJobs chan models.CategoryJob, results chan *os.File) {
	// 	wg.Wait()
	// 	close(results)
	// }(writeWg, fileJobs, fileResults)

	// for fr := range fileResults {
	// 	fmt.Println(fr.Name())
	// 	fr.Close()
	// }

	// log.Println("Finished")
}
