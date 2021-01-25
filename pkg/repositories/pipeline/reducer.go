package pipeline

import (
	"log"
	"os"
)

// Reducer represents implementation of Pipe for resulting data
type Reducer struct {
}

// NewReducer returns representation of Pipe
func NewReducer() Pipe {
	return Reducer{}
}

// Call implements call of the Pipe
func (c Reducer) Call(in, out chan interface{}) {
	categoryFiles := make(map[string]*os.File)
	for data := range in {
		wr := data.(WriteResult)
		categoryFiles[wr.Category] = wr.File
	}

	for category, file := range categoryFiles {
		log.Printf("Category: %s; File: %s", category, file.Name())
		file.Sync()
		file.Close()
	}
}
