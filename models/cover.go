package models

type Cover struct {
	ID          uint   `gorm:"primary_key"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	ContentType string `json:"content-type"`
}

/*type Cover struct {
  ID          uint      `gorm:"primary_key"`
  Name        string    `json:"name"`
  Value       string    `json:"value"`
  ContentType string    `json:"content-type"`
}*/
