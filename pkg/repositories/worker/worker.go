package worker

import (
	"fmt"
	"log"
	"net/http"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"sync"
	"time"
)

type Work struct {
	ID     int
	Record *models.Record
}

type Worker struct {
	ID            int
	WorkerChannel chan chan Work
	Channel       chan Work
	End           chan bool
	ResultData    chan *models.ResultData
	Errors        chan error
	client        *http.Client
	handler       handler.HandlerInt
	wg            *sync.WaitGroup
}

func NewWorker(
	id int,
	workerChannel chan chan Work,
	channel chan Work,
	end chan bool,
	handler handler.HandlerInt,
	results chan *models.ResultData,
	errorsChan chan error,
	wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:            id,
		Channel:       channel,
		WorkerChannel: workerChannel,
		End:           end,
		ResultData:    results,
		Errors:        errorsChan,
		handler:       handler,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		wg: wg,
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerChannel <- w.Channel
			select {
			case work := <-w.Channel:
				result, _ := w.fetchPage(work.ID, work.Record)
				fmt.Println(result)
				// if fetchErr != nil {
				// 	w.Errors <- fetchErr
				// } else {
				w.ResultData <- result
				// }
			case <-w.End:
				w.wg.Done()
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	log.Printf("worker [%d] is stopping", w.ID)
	w.End <- true
}

func (w *Worker) fetchPage(id int, record *models.Record) (*models.ResultData, error) {
	resp, respErr := w.client.Get(record.URL)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		data := &models.ResultData{
			Title:       "Not Found",
			Description: "Not Found",
			URL:         record.URL,
			Categories:  record.Categories,
		}
		return data, nil
	}
	resultData, resultErr := w.handler.Parse(resp.Body)
	if resultErr != nil {
		return nil, resultErr
	}
	resultData.URL = record.URL
	resultData.Categories = record.Categories
	return resultData, nil
}
