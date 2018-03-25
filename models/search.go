package models

// Search data model
type Search struct {
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Genre      string   `json:"genre"`
	Section    string   `json:"section"`
	SubSection string   `json:"subsection"`
	Series     string   `json:"series"`
	Ser        string   `json:"ser"`
	SearchSer  string   `json:"search_ser"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	Deleted    bool     `json:"deleted"`
	Langs      []string `json:"langs"`
}
