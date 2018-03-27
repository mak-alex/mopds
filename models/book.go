package models

import (
	"regexp"

	"bitbucket.org/enlab/mopds/utils"
)

var re = regexp.MustCompile(`^[\s\,]*(.*?)[\,\s]*$`)

type Book struct {
  ID          uint         `json:"id" gorm:"primary_key"`
	CatalogID   uint         `json:"-"`
	FileName    string       `json:"filename" gorm:"unique_index"`
	Path        string       `json:"path"`
	Format      string       `json:"format"`
	Title       string       `json:"title"`
	DocDate     string       `json:"docdate"`
	Lang        string       `json:"lang"`
	FileSize    string       `json:"filesize"`
	Rating      uint         `json:"rating"`
	Rank        uint         `json:"rank"`
	Avail       uint         `json:"avail"`
	LibID       string       `json:"lib_id" gorm:"index"`
	CatalogType uint         `json:"cat_type"`
	Catalog     Catalog      `json:"catalog"`
	LangCode    int          `json:"lang_code"`
	SerNo       string       `json:"ser_no,omitempty"`
	Del         string       `json:"del"`
	Authors     []Author     `json:"authors" gorm:"many2many:bauthors;"`
	Annotations []Annotation `json:"annotations" gorm:"many2many:bannotations;"`
	Genres      []Genre      `json:"genres"  gorm:"many2many:bgenres;"`
	Covers      []Cover      `json:"covers" gorm:"many2many:bcovers;"`
	Series      []Serie      `json:"series" gorm:"many2many:bseries;"`
	Update      bool         `json:"-" gorm:"-"`
}

func (book *Book) AfterFind() {
	book.Title = utils.UpperInitialAll(book.Title)
}

func (book *Book) GetFullFilename() string {
	authors := ""
	for _, a := range book.Authors {
		authors = authors + a.FullName
	}
	authors = re.ReplaceAllString(authors, "$1")
	authRunes := []rune(authors)
	if len(authRunes) > 100 {
		authors = string(authRunes[0:100]) + `…`
	}
	outName := authors + " - "
	/*if book.SerNo != "" && book.SerNo != "0" {
		if len(book.SerNo) == 1 {
			book.SerNo = "0" + book.SerNo
		}
		outName = outName + "[" + book.SerNo + "] "
	}*/

	return outName + book.Title + "." + book.Format
}
