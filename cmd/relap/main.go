package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
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

func decodeLine(bytes []byte) (*models.Record, error) {
	var record models.Record
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}

type Reader struct {
	file    *os.File
	results chan models.Result
	wg      *sync.WaitGroup
	jobs    chan models.Job
	errors  chan error
}

type Writer struct {
	wg            *sync.WaitGroup
	jobs          chan worker.PoolWriteJob
	results       chan worker.WriteResult
	errors        chan error
	mu            *sync.Mutex
	categoryFiles map[string]*os.File
}

type Combiner struct {
}

func (r Reader) Call(in, out chan interface{}) {
	mainWg := &sync.WaitGroup{}
	recordWg := &sync.WaitGroup{}
	go func(file *os.File, jobs chan models.Job, errors chan error, wg *sync.WaitGroup, mainWg *sync.WaitGroup) {
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
				jobs <- models.Job{Record: record}
			}
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			errors <- scannerErr
		}
	}(r.file, r.jobs, r.errors, recordWg, mainWg)

	go func(wg *sync.WaitGroup, results chan models.Result) {
		wg.Wait()
		close(results)
	}(r.wg, r.results)

	for res := range r.results {
		if res.Err == nil {
			out <- res.Result
		}
	}

}

func (w Writer) Call(in, out chan interface{}) {
	go func(in chan interface{}, w *Writer) {
		defer close(w.jobs)
		for data := range in {
			resultData := data.(*models.ResultData)
			for _, category := range resultData.Categories {
				var (
					catFile *os.File
				)
				catFile = w.getCategoryFile(category)
				if catFile == nil {
					catFile, _ = w.setCategoryFile(category)
				}
				w.jobs <- worker.PoolWriteJob{File: catFile, ResultData: resultData, Category: category}
			}
		}
	}(in, &w)

	go func(wg *sync.WaitGroup, results chan worker.WriteResult) {
		wg.Wait()
		close(results)
	}(w.wg, w.results)

	for res := range w.results {
		out <- res
	}
}

func (c Combiner) Call(in, out chan interface{}) {
	categoryFiles := make(map[string]*os.File)
	for data := range in {
		wr := data.(worker.WriteResult)
		categoryFiles[wr.Category] = wr.File
	}

	for category, file := range categoryFiles {
		fmt.Println("Category: ", category, "File: ", file)
		file.Sync()
		file.Close()
	}
}

func (w Writer) getCategoryFile(category string) *os.File {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.categoryFiles[category]
}

func (w Writer) setCategoryFile(category string) (*os.File, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	categoryFile, err := os.OpenFile(category+".tsv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error of creating %s file: %s", category, err)
	}
	return categoryFile, nil
}

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
	// recordFile := record.NewFile(
	// 	wg,
	// 	fs,
	// 	jobs,
	// 	results,
	// 	errors,
	// )
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

	pipes := []pipeline.Pipe{
		Reader{
			file:    file,
			results: results,
			jobs:    jobs,
			errors:  errors,
			wg:      wg,
		},
		Writer{
			wg:            writeWg,
			jobs:          pwj,
			results:       writeResults,
			categoryFiles: make(map[string]*os.File),
			mu:            &sync.Mutex{},
		},
		Combiner{},
	}
	pipeline.ExecutePipeline(pipes...)
	elapsed := time.Since(startTime)
	log.Printf("Program has finished. It took %s", elapsed)

	// 	readErr := File.ReadLines(file)
	// 	if readErr != nil {
	// 		log.Fatal(readErr)
	// 	}
	// 	// workersPool.ReadFromChannels(results, pwj, errors)
	// 	_, readFileError := workersPool.ReadFromChannels(results, pwj, errors)
	// 	if readFileError != nil {
	// 		log.Fatalf("Error file reading: %s", readFileError.Error())
	// 	}
	// 	// go func(wg *sync.WaitGroup, pwj chan worker.PoolWriteJob) {
	// 	// 	wg.Done()
	// 	// 	// close(pwj)
	// 	// }(writeWg, pwj)

	// 	var categoriesFiles = make(map[string]*os.File)
	// WRITE:
	// 	for {
	// 		select {
	// 		case wr, ok := <-writeResults:
	// 			if !ok {
	// 				break WRITE
	// 			}
	// 			fmt.Println(wr, ok)
	// 			_, exists := categoriesFiles[wr.Category]
	// 			if !exists {
	// 				categoriesFiles[wr.Category] = wr.File
	// 			}
	// 		}
	// 	}
	// 	for _, f := range categoriesFiles {
	// 		syncErr := f.Sync()
	// 		if syncErr != nil {
	// 			fmt.Println(syncErr)
	// 		}
	// 		f.Close()
	// 	}
	// 	fmt.Println("Finished")
}
