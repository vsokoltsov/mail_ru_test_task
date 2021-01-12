package worker

import (
	"relap/pkg/models"
)

// Int represents actions for worker interface
type Int interface {
	FetchPage(url string, categories []string) (*models.ResultData, error)
}
