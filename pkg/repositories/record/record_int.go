package record

import (
	"os"
	"relap/pkg/models"
)

type RecordInt interface {
	ReadLines(file *os.File) error
	decodeLine(data []byte) (*models.Record, error)
}
