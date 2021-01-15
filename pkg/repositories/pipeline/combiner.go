package pipeline

import (
	"fmt"
	"os"
	"relap/pkg/models"
)

type Combiner struct {
}

func NewCombiner() Pipe {
	return Combiner{}
}

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
