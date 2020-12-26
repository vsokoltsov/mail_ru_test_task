package worker

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"relap/pkg/models"
	"relap/pkg/repositories/handler"
	"strings"
	"testing"
)

type workerTestCase struct {
	name           string
	handler        handler.Int
	expectErr      bool
	generateServer func() *httptest.Server
	url            string
}

type MockSuccessHandler struct {
}
type MockFailedHandler struct {
}

func (mshi MockSuccessHandler) Parse(body io.ReadCloser) (*models.ResultData, error) {
	return &models.ResultData{
		Title:       "Title",
		Description: "Description",
	}, nil
}

func (mfhi MockFailedHandler) Parse(body io.ReadCloser) (*models.ResultData, error) {
	return nil, fmt.Errorf("Parse error")
}

var testCases = []workerTestCase{
	workerTestCase{
		name:    "Success FetchPage",
		handler: MockSuccessHandler{},
		generateServer: func() *httptest.Server {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Send response to be tested
				rw.Write([]byte(`{ "success": "true" }`))
			}))
			return server
		},
	},
	workerTestCase{
		name:    "Failed FetchPage (client error)",
		handler: MockSuccessHandler{},
		generateServer: func() *httptest.Server {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Send response to be tested
				rw.Write([]byte(`{ "success": "true" }`))
			}))
			return server
		},
		url:       "test",
		expectErr: true,
	},
	workerTestCase{
		name:    "Failed FetchPage (page not found)",
		handler: MockSuccessHandler{},
		generateServer: func() *httptest.Server {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Send response to be tested
				rw.WriteHeader(http.StatusNotFound)
			}))
			return server
		},
	},
	workerTestCase{
		name:    "Failed FetchPage (page not found)",
		handler: MockFailedHandler{},
		generateServer: func() *httptest.Server {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Send response to be tested
				rw.WriteHeader(http.StatusOK)
			}))
			return server
		},
		expectErr: true,
	},
}

func TestWorker(t *testing.T) {
	for _, tc := range testCases {
		label := strings.Join([]string{"Repo", "Worker", tc.name}, " ")
		t.Run(label, func(t *testing.T) {
			var url string
			server := tc.generateServer()
			worker := Worker{
				client:  server.Client(),
				handler: tc.handler,
			}
			if tc.url != "" {
				url = tc.url
			} else {
				url = server.URL
			}
			_, fetchPageErr := worker.FetchPage(url, []string{"cat1"})
			if tc.expectErr && fetchPageErr == nil {
				t.Errorf("Expected error, got nil")
			} else if !tc.expectErr && fetchPageErr != nil {
				t.Errorf("Unexpect error: %s", fetchPageErr)
			}
		})
	}
}

func TestNewWorker(t *testing.T) {
	w := NewWorker(MockSuccessHandler{})
	if w == nil {
		t.Errorf("Initializatio have not happened")
	}
}
