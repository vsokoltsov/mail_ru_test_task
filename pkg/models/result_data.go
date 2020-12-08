package models

type Job struct {
	Record *Record
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
