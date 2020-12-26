package handler

import (
	"io"
	"relap/pkg/models"
)

// Int represents available actions for handler
type Int interface {
	Parse(body io.ReadCloser) (*models.ResultData, error)
}
