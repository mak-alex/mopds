package models

type Genre struct {
	ID         int   `json:"id" gorm:"primary_key"`
	Genre      string `json:"genre_code" gorm:"not null;"`
	Section    string `json:"section" gorm:"not null;"`
	Subsection string `json:"subsection" gorm:"not null;"`
}
