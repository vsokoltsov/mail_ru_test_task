package models

import "os"

type Job struct {
	Record *Record
}

type CategoryJob struct {
	Category    string
	ResultsData []*ResultData
	File        *os.File
}

type Result struct {
	WorkerID int
	Result   *ResultData
	Err      error
}
type ResultData struct {
	Title       string
	Description string
	URL         string
	Categories  []string
}
