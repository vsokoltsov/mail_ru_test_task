package handler

import (
	"io"
)

// Int represents available actions for handler
type Int interface {
	Parse(body io.ReadCloser) (*ResultData, error)
}
