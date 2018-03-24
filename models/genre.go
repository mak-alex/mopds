package models

type Genre struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	Genre string `json:"genre_code" gorm:"not null;unique_index"`
	Section   string `json:"section" gorm:"not null;unique_index"`
  Subsection string `json:"subsection" gorm:"not null;unique_index"`
}
