package handler

import (
	"io"
	"relap/pkg/models"
)

type HandlerInt interface {
	Parse(body io.ReadCloser) (*models.ResultData, error)
}
