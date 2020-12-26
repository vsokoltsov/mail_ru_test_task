package handler

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

type handlerTestCase struct {
	name                string
	html                []byte
	expectedTitle       string
	expectedDescription string
	expectedError       bool
}

var testCases = []handlerTestCase{
	handlerTestCase{
		name: "Success parsing of file",
		html: []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Test</title>
				<meta name="description" content="Test description"/>
			</head>
			<body>
				body content
				<p>more content</p>
			</body>
			</html>
		`),
		expectedDescription: "Test description",
		expectedTitle:       "Test",
	},
	handlerTestCase{
		name: "Success parsing of file (title is in meta)",
		html: []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta name="description" content="Test description"/>
				<meta name="title" content="title in meta"/>
			</head>
			<body>
				body content
				<p>more content</p>
			</body>
			</html>
		`),
		expectedDescription: "Test description",
		expectedTitle:       "title in meta",
	},
	handlerTestCase{
		name: "Success parsing of file (different description name)",
		html: []byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta name="Description" content="Capital letter"/>
				<meta name="title" content="title in meta"/>
			</head>
			<body>
				body content
				<p>more content</p>
			</body>
			</html>
		`),
		expectedDescription: "Capital letter",
		expectedTitle:       "title in meta",
	},
	handlerTestCase{
		name: "Failed parsing of file (break case)",
		html: []byte(`
			awdaw
		`),
		expectedDescription: "",
		expectedTitle:       "",
	},
	handlerTestCase{
		name:                "Failed parsing of file (html.TokenError)",
		html:                []byte(`error`),
		expectedDescription: "",
		expectedTitle:       "",
	},
}

func TestHandlers(t *testing.T) {
	for _, tc := range testCases {
		label := strings.Join([]string{"Handler", tc.name}, " ")
		t.Run(label, func(t *testing.T) {
			r := ioutil.NopCloser(bytes.NewReader(tc.html))

			handler := NewHTML()
			result, err := handler.Parse(r)
			if tc.expectedError {
				if err == nil {
					t.Errorf("Error of parsing the content: %s", err)
				}
			} else {
				if result == nil {
					t.Errorf("Result of parsing is nul")
				}

				if result.Title != tc.expectedTitle {
					t.Errorf("Title does not match: %s and %s", result.Title, tc.expectedTitle)
				}

				if result.Description != tc.expectedDescription {
					t.Errorf("Description does not match: %s and %s", result.Description, tc.expectedDescription)
				}
			}
		})
	}
}
