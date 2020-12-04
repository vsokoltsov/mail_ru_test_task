package worker

import (
	"log"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
)

var WorkerChannel = make(chan chan Work)

type Collector struct {
	Work chan Work
	End  chan bool
}

func StartDispatcher(workerCount int, handler handler.HandlerInt, resultsChan chan *models.ResultData, errorsChan chan error) Collector {
	var i int
	var workers []*Worker
	input := make(chan Work)
	end := make(chan bool)
	collector := Collector{Work: input, End: end}

	for i = 1; i <= workerCount; i++ {
		log.Println("starting worker: ", i)
		worker := NewWorker(i, WorkerChannel, make(chan Work), make(chan bool), handler, resultsChan, errorsChan)
		worker.Start()
		workers = append(workers, worker)
	}

	go func() {
		for {
			select {
			case <-end:
				for _, w := range workers {
					w.Stop()
				}
				return
			case work := <-input:
				worker := <-WorkerChannel
				worker <- work
			}
		}
	}()

	return collector
}
