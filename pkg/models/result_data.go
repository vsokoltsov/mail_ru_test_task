package models

import "os"

// Job defines job for read workers pool
type Job struct {
	Record *Record
}

// CategoryJob defines job for write workers pool
type CategoryJob struct {
	Category    string
	ResultsData []*ResultData
	File        *os.File
}

// Result represents operation outcome
type Result struct {
	WorkerID int
	Result   *ResultData
	Err      error
}

// ResultData contains final information for file writing
type ResultData struct {
	Title       string
	Description string
	URL         string
	Categories  []string
}
