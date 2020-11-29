package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/net/html"
)

type Record struct {
	URL             string   `json:"url"`
	State           string   `json:"state"`
	Categories      []string `json:"categories"`
	CategoryAnother string   `json:"category_another"`
	ForMainPage     bool     `json:"for_main_page"`
	Ctime           int      `json:"ctime"`
}

type FileData struct {
	Title       string
	Description string
	URL         string
	Categories  []string
}

func getRequest(url string) ([]byte, error) {
	res, _ := http.Get(url)

	return ioutil.ReadAll(res.Body)
}

func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
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

func prepareString(str string) string {
	data := strings.Replace(str, "\n", " ", -1)
	data = strings.TrimSpace(data)
	if !utf8.ValidString(data) {
		log.Printf("Non-utf8 encoding: %s", data)
	}
	return data
}

func fetchPage(url string, categories []string) *FileData {
	var (
		data = FileData{
			URL:        url,
			Categories: categories,
		}
	)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Get request error: %s", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		data.Title = "Not Found"
		data.Description = "Not Found"
		return &data
	}
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		t := tokenizer.Token()
		tokenErr := tokenizer.Err()
		if tokenErr == io.EOF {
			break
		}

		switch tt {
		case html.ErrorToken:
			log.Fatal(html.ErrorToken)
		case html.StartTagToken, html.SelfClosingTagToken:
			if t.Data == "title" {
				tokenType := tokenizer.Next()
				if tokenType == html.TextToken {
					data.Title = prepareString(tokenizer.Token().Data)
					break
				}
			}
			if t.Data == "meta" {
				metaDesc, isDesc := extractMetaProperty(t, "description")
				if isDesc {
					data.Description = prepareString(metaDesc)
				}
				metaDescUp, isDescUp := extractMetaProperty(t, "Description")
				if isDescUp {
					data.Description = prepareString(metaDescUp)
				}
				metaTitle, isTitle := extractMetaProperty(t, "title")
				if isTitle && data.Title != "" {
					data.Title = prepareString(metaTitle)
				}
			}
		}
		if data.Title != "" && data.Description != "" {
			break
		}
	}
	return &data
}

func fetchPages(records []*Record, urlsChan chan *FileData, globalWG *sync.WaitGroup) {
	defer globalWG.Done()

	for _, rec := range records {
		d := fetchPage(rec.URL, rec.Categories)
		urlsChan <- d
	}

}

func writeToFile(
	categoryName string,
	fileDatas []*FileData) {

	var (
		f       *os.File
		fileErr error
	)

	path := "./results/" + categoryName
	if _, err := os.Stat(path); err == nil {
		f, fileErr = os.Open(path)
	} else {
		f, fileErr = os.Create("./results/" + categoryName + ".tsv")
	}
	defer f.Close()

	if fileErr != nil {
		log.Fatal(fileErr)
	}

	for _, fd := range fileDatas {
		f.Write([]byte(strings.Join([]string{fd.URL, fd.Title, fd.Description, "\n"}, " ")))
	}

	f.Sync()
}

func main() {
	var (
		goroutinesGroup = 25
		wg              = sync.WaitGroup{}
		urlsChan        = make(chan *FileData)
		counter         int
		records         []*Record
		categoriesData  = make(map[string][]*FileData)
	)

	file, err := os.Open("./500.jsonl")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		var record Record
		bytes := scanner.Bytes()
		if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
			log.Fatalf("Error of record unmarshalling: %s", unmarshalErr)
		}

		counter++
		records = append(records, &record)
		if counter == goroutinesGroup {
			wg.Add(1)
			go fetchPages(records, urlsChan, &wg)
			counter = 0
			records = []*Record{}
		}
	}

	if scannerErr := scanner.Err(); scannerErr != nil {
		log.Fatal(scannerErr)
	}

	// It works, because of the goroutine scheduler
	go func(wg *sync.WaitGroup, urlsChan chan *FileData) {
		wg.Wait()
		close(urlsChan)
	}(&wg, urlsChan)

	for fileData := range urlsChan {
		if fileData != nil {
			for _, category := range fileData.Categories {
				_, ok := categoriesData[category]
				if ok {
					categoriesData[category] = append(categoriesData[category], fileData)
				} else {
					categoriesData[category] = []*FileData{fileData}
				}
			}
		}
	}

	for key, value := range categoriesData {
		writeToFile(key, value)
	}

	fmt.Println("Finished execution")
}
