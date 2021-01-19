package pipeline

import (
	"fmt"
	"os"
	"relap/pkg/models"
)

// Combiner represents implementation of Pipe for resulting data
type Combiner struct {
}

// NewCombiner returns representation of Pipe
func NewCombiner() Pipe {
	return Combiner{}
}

// Call implements call of the Pipe
func (c Combiner) Call(in, out chan interface{}) {
	categoryFiles := make(map[string]*os.File)
	for data := range in {
		wr := data.(models.WriteResult)
		categoryFiles[wr.Category] = wr.File
	}

	for category, file := range categoryFiles {
		fmt.Println("Category:", category, "File:", file.Name())
		file.Sync()
		file.Close()
	}
}
