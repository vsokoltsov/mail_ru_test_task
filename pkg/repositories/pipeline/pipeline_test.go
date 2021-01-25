package pipeline

import (
	"sync/atomic"
	"testing"
)

var entries uint32

type Step1 struct{}

func (s1 Step1) Call(in, out chan interface{}) {
	out <- 1
	out <- 2
	out <- 3
}

type Step2 struct{}

func (s2 Step2) Call(in, out chan interface{}) {
	for range in {
		atomic.AddUint32(&entries, 1)
	}
}

func TestPipeline(t *testing.T) {
	pipeline := []Pipe{
		Step1{},
		Step2{},
	}
	ExecutePipeline(pipeline...)
	if entries != 3 {
		t.Errorf("Execution of pipeline has failed")
	}
}
