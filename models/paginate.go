package models

// PaginatedList represents a paginated list of data items.
type PaginatedList struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	PageCount  int         `json:"page_count"`
	TotalCount int         `json:"total_count"`
	Items      interface{} `json:"items"`
}
