package models

type Catalog struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	CatName  string `json:"cat_name"`
	Path     string `json:"path"`
	CatType  uint   `json:"cat_type"`
	CatSize  int64  `json:"cat_size"`
	ParentId uint   `json:"parent_id"`
}
