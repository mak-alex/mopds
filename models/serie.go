package models

type Serie struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	Ser       string `json:"ser"`
	SearchSer string `json:"search_ser"`
}
