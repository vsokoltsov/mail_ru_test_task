package record

import (
	"os"
	"relap/pkg/models"
)

type RecordInt interface {
	ReadLines(file *os.File) (map[string][]*models.ResultData, error)
	SaveResults(dir, ext, categoryName string, results []*models.ResultData) (string, error)
	decodeLine(data []byte) (*models.Record, error)
	fetchPages(records []*models.Record)
	fetchPage(url string, categories []string) (*models.ResultData, error)
}
