package handler

import (
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// HTML implements handler.Int interface
type HTML struct{}

// ResultData contains final information for file writing
type ResultData struct {
	Title       string
	Description string
	URL         string
	Categories  []string
}

// NewHTML returns new HTML instance
func NewHTML() Int {
	return &HTML{}
}

// extractMetaProperty extracts attributes based on parameters
func (hh *HTML) extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "name" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}

	return
}

// Parse parsed give html page
func (hh *HTML) Parse(body io.ReadCloser) (*ResultData, error) {
	result := ResultData{}
	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		t := tokenizer.Token()
		tokenErr := tokenizer.Err()
		if tokenErr == io.EOF {
			break
		}

		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			if t.Data == "title" {
				tokenType := tokenizer.Next()
				if tokenType == html.TextToken {
					result.Title = prepareString(tokenizer.Token().Data)
					break
				}
			}
			if t.Data == "meta" {
				metaDesc, isDesc := hh.extractMetaProperty(t, "description")
				if isDesc {
					result.Description = prepareString(metaDesc)
				}
				metaDescUp, isDescUp := hh.extractMetaProperty(t, "Description")
				if isDescUp {
					result.Description = prepareString(metaDescUp)
				}
				metaTitle, isTitle := hh.extractMetaProperty(t, "title")
				if isTitle && result.Title == "" {
					result.Title = prepareString(metaTitle)
				}
			}
		}
		if result.Title != "" && result.Description != "" {
			break
		}
	}
	return &result, nil
}

func prepareString(str string) string {
	data := strings.Replace(str, "\n", " ", -1)
	data = strings.TrimSpace(data)
	if !utf8.ValidString(data) {
		log.Printf("Non-utf8 encoding: %s", data)
	}
	return data
}
