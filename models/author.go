package models

type Author struct {
	ID              uint      `gorm:"primary_key"`
	FullName        string    `json:"full_name"`
	SearchFullName  string    `json:"search_full_name"`
  LangCode        string    `json:"lang_code"`
}
