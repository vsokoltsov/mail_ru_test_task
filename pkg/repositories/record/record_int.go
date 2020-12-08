package record

import (
	"os"
	"relap/pkg/models"
)

type RecordInt interface {
	ReadLines(file *os.File) error
	SaveResults(dir, ext, categoryName string, results []*models.ResultData) (string, error)
	decodeLine(data []byte) (*models.Record, error)
}
