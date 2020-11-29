package main

import (
	"flag"
	"log"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"relap/pkg/repositories/record"
	"relap/pkg/repositories/storage"
	"sync"
)

func main() {
	var (
		path           = flag.String("file", "./../../500.jsonl", "Path to a file")
		resultsDir     = flag.String("results", "./../../results/", "Folder with result files.")
		resultExt      = flag.String("ext", "tsv", "Extension of the resulting file.")
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
		fs,
	)

	file, fileErr := fs.OpenFile(*path)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	defer file.Close()

	results, readErr := recordFile.ReadLines(file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	for category, results := range results {
		path, saveErr := recordFile.SaveResults(*resultsDir, *resultExt, category, results)
		if saveErr != nil {
			log.Fatal(saveErr)
		}
		log.Printf("Path for %s category: %s", category, path)
	}
}
