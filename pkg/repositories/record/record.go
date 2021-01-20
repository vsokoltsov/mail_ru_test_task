package record

import (
	"encoding/json"
	"fmt"
)

// Row represents row in file
type Row struct {
	URL             string   `json:"url"`
	State           string   `json:"state"`
	Categories      []string `json:"categories"`
	CategoryAnother string   `json:"category_another"`
	ForMainPage     bool     `json:"for_main_page"`
	Ctime           int      `json:"ctime"`
}

func DecodeLine(bytes []byte) (*Row, error) {
	var record Row
	if unmarshalErr := json.Unmarshal(bytes, &record); unmarshalErr != nil {
		return nil, fmt.Errorf("Error of record unmarshalling: %s", unmarshalErr)
	}
	return &record, nil
}
