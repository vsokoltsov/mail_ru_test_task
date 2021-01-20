package worker

import (
	"net/http"
	"relap/pkg/repositories/handler"
	"time"
)

// Worker represents worker info
type Worker struct {
	client  *http.Client
	handler handler.Int
}

// NewWorker returns structure that implements Int
func NewWorker(handler handler.Int) Int {
	return Worker{
		handler: handler,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchPage fetch page based on url and write result to channel
func (w Worker) FetchPage(url string, categories []string) (*handler.ResultData, error) {
	resp, respErr := w.client.Get(url)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		data := &handler.ResultData{
			Title:       "Not Found",
			Description: "Not Found",
			URL:         url,
			Categories:  categories,
		}
		return data, nil
	}
	resultData, resultErr := w.handler.Parse(resp.Body)
	if resultErr != nil {
		return nil, resultErr
	}
	resultData.URL = url
	resultData.Categories = categories
	return resultData, nil
}
