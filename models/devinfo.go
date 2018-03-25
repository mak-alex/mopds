package models

type DevInfo struct {
	Author  string `json:"author"`
	Email   string `json:"email"`
	Project struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Link    string `json:"link"`
		Created string `json:"created"`
	} `json:"project"`
}
