package models

type Summary struct {
	Authors  int `json:"authors"`
	Books    int `json:"books"`
	Catalogs int `json:"catalogs"`
	Genres   int `json:"genres"`
	Series   int `json:"series"`
}
