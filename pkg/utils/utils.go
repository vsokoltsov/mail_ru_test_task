package utils

import (
	"log"
	"strings"
	"unicode/utf8"
)

func PrepareString(str string) string {
	data := strings.Replace(str, "\n", " ", -1)
	data = strings.TrimSpace(data)
	if !utf8.ValidString(data) {
		log.Printf("Non-utf8 encoding: %s", data)
	}
	return data
}
