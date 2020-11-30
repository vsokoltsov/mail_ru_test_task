package handler

import (
	"io"
	"relap/pkg/models"
	"relap/pkg/utils"

	"golang.org/x/net/html"
)

type HandlerHtml struct{}

func NewHandlerHTML() HandlerInt {
	return &HandlerHtml{}
}

func (hh *HandlerHtml) extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
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

func (hh *HandlerHtml) Parse(body io.ReadCloser) (*models.ResultData, error) {
	result := models.ResultData{}
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
					result.Title = utils.PrepareString(tokenizer.Token().Data)
					break
				}
			}
			if t.Data == "meta" {
				metaDesc, isDesc := hh.extractMetaProperty(t, "description")
				if isDesc {
					result.Description = utils.PrepareString(metaDesc)
				}
				metaDescUp, isDescUp := hh.extractMetaProperty(t, "Description")
				if isDescUp {
					result.Description = utils.PrepareString(metaDescUp)
				}
				metaTitle, isTitle := hh.extractMetaProperty(t, "title")
				if isTitle && result.Title == "" {
					result.Title = utils.PrepareString(metaTitle)
				}
			}
		}
		if result.Title != "" && result.Description != "" {
			break
		}
	}
	return &result, nil
}
