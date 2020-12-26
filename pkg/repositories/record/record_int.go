package record

import (
	"os"
	"relap/pkg/models"
)

// Int represents available actions for record
type Int interface {
	ReadLines(file *os.File) error
	decodeLine(data []byte) (*models.Record, error)
}
