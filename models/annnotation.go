package models

type Annotation struct {
	ID    uint   `json:"-" gorm:"primary_key"`
	Value string `json:"value" gorm:"not null;"`
}
