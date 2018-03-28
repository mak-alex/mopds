package models

type SubItemMenu struct {
	ID      string `json:"id"`
	GenreID int    `json:"genre_id"`
	Value   string `json:"value"`
}

type ItemMenu struct {
	ID    string        `json:"id"`
	Value string        `json:"value"`
	Data  []SubItemMenu `json:"data"`
}
