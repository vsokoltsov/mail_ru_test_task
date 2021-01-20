package worker

import "relap/pkg/repositories/handler"

// Int represents actions for worker interface
type Int interface {
	FetchPage(url string, categories []string) (*handler.ResultData, error)
}
