package pipeline

import "sync"

type Pipe interface {
	Call(in, out chan interface{})
}

func ExecutePipeline(pipes ...Pipe) {
	in := make(chan interface{}, 1)
	wg := &sync.WaitGroup{}
	wg.Add(len(pipes))
	for idx := range pipes {
		out := make(chan interface{}, 1)
		j := pipes[idx]
		go executeJob(in, out, j, wg)
		in = out
	}
	wg.Wait()
}

func executeJob(in, out chan interface{}, j Pipe, wg *sync.WaitGroup) {
	defer wg.Done()
	j.Call(in, out)
	close(out)
}
