package models

// Search data model
type Search struct {
	Title   string   `json:"title"`
	Author  string   `json:"author"`
	Genre   string   `json:"genre"`
	Series  string   `json:"serie"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
	Page    int      `json:"page"`
	PerPage int      `json:"per_page"`
	Deleted bool     `json:"deleted"`
	Langs   []string `json:"langs"`
}
