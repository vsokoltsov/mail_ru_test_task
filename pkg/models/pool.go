package models

import "os"

// ReadJob defines job for read workers pool
type ReadJob struct {
	Record *Record
}

type WriteJob struct {
	WorkerID   int
	File       *os.File
	ResultData *ResultData
	Category   string
}

// Result represents operation outcome
type ReadResult struct {
	WorkerID int
	Result   *ResultData
	Err      error
}

type WriteResult struct {
	Category string
	File     *os.File
}
