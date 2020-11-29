package models

type Record struct {
	URL             string   `json:"url"`
	State           string   `json:"state"`
	Categories      []string `json:"categories"`
	CategoryAnother string   `json:"category_another"`
	ForMainPage     bool     `json:"for_main_page"`
	Ctime           int      `json:"ctime"`
}
