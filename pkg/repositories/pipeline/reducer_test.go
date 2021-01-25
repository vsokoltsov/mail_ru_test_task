package pipeline

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestReducerDataRead(t *testing.T) {
	var (
		in             = make(chan interface{}, 1)
		out            = make(chan interface{})
		exampleFile, _ = os.OpenFile("test.tsv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	)
	var buf bytes.Buffer
	log.SetOutput(&buf)

	defer os.Remove("test.tsv")

	reducer := NewReducer()
	in <- WriteResult{
		File:     exampleFile,
		Category: "test",
	}
	close(in)
	reducer.Call(in, out)
	res := buf.String()
	if !strings.Contains(res, "test") {
		t.Errorf("Wrong logging message")
	}
}
