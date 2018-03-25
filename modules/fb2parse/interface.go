package fb2

import (
	"encoding/xml"

	proto "bitbucket.org/enlab/mopds/modules/fb2parse/prototype"
)

// List of interfaces for integration

// FB2 represents FB2 structure
//proteus:generate
type FB2 struct {
	ID          string   `json:"_id"`
	FictionBook xml.Name `xml:"FictionBook" json:"FictionBook"`
	Stylesheet  []string `xml:"stylesheet" json:"stylesheet"`
	Description struct {
		TitleInfo struct {
			Genre      []string     `xml:"genre" json:"genre"`
			GenreType  []string     `xml:"genreType" json:"genreType"`
			Author     []AuthorType `xml:"author" json:"author"`
			BookTitle  string       `xml:"book-title" json:"book-title"`
			Annotation []string     `xml:"annotation>p" json:"annotation"`
			Keywords   string       `xml:"keywords" json:"keywords"`
			Date       string       `xml:"date" json:"date"`
			Coverpage  struct {
				Image struct {
					Href string `xml:"xlink:href,attr" json:"href"`
				} `xml:"image,allowempty" json:"image"`
			} `xml:"coverpage" json:"coverpage"`
			Lang       string     `xml:"lang" json:"lang"`
			SrcLang    string     `xml:"src-lang" json:"src-lang"`
			Translator AuthorType `xml:"translator" json:"translator"`
			Sequence   string     `xml:"sequence" json:"sequence"`
		} `xml:"title-info" json:"title-info"`
		DocumentInfo struct {
			Author      []AuthorType `xml:"author" json:"author"`
			ProgramUsed string       `xml:"program-used" json:"program-used"`
			Date        string       `xml:"date" json:"date"`
			SrcURL      []string     `xml:"src-url" json:"src-url"`
			SrcOcr      string       `xml:"src-ocr" json:"src-ocr"`
			ID          string       `xml:"id" json:"id"`
			Version     float64      `xml:"version" json:"version"`
			History     string       `xml:"history" json:"history"`
		} `xml:"document-info" json:"document-info"`
		PublishInfo struct {
			BookName  string `xml:"book-name" json:"book-name"`
			Publisher string `xml:"publisher" json:"publisher"`
			City      string `xml:"city" json:"city"`
			Year      int    `xml:"year" json:"year"`
			ISBN      string `xml:"isbn" json:"isbn"`
			Sequence  string `xml:"sequence" json:"sequence"`
		} `xml:"PublishInfo" json:"PublishInfo"`
		CustomInfo []struct {
			InfoType string `xml:"info-type" json:"info-type"`
		} `xml:"custom-info" json:"custom-info"`
	} `xml:"description" json:"description"`
	Body struct {
		Sections []struct {
			P []string `xml:"p" json:"p"`
		} `xml:"section" json:"section"`
	} `xml:"body" json:"body"`
	Binary []struct {
		Value       string `xml:",chardata" json:"value"`
		ContentType string `xml:"content-type,attr" json:"content-type"`
		ID          string `xml:"id,attr" json:"id"`
	} `xml:"binary" json:"binary"`
}

// UnmarshalCoverpage func
func (f *FB2) UnmarshalCoverpage(data []byte) {
	tagOpened := false
	coverpageStartIndex := 0
	coverpageEndIndex := 0
	// imageHref := ""
	tagName := ""
_loop:
	for i, v := range data {
		if tagOpened {
			switch v {
			case '>':
				if tagName != "p" && tagName != "/p" {
				}
				tagOpened = false
				if tagName == "coverpage" {
					coverpageStartIndex = i + 1
				} else if tagName == "/coverpage" {
					coverpageEndIndex = i - 11
					break _loop
				}
				tagName = ""
				break
			default:
				tagName += string(v)
			}
		} else {
			if v == '<' {
				tagOpened = true
			}
		}
	}

	if coverpageEndIndex > coverpageStartIndex {
		href := parseImage(data[coverpageStartIndex:coverpageEndIndex])
		f.Description.TitleInfo.Coverpage.Image.Href = href
	}
}

// AuthorType embedded fb2 type, represents author info
type AuthorType struct {
	FirstName  string `xml:"first-name"`
	MiddleName string `xml:"middle-name"`
	LastName   string `xml:"last-name"`
	Nickname   string `xml:"nickname"`
	HomePage   string `xml:"home-page"`
	Email      string `xml:"email"`
}

// TextFieldType embedded fb2 type, represents text field
type TextFieldType struct {
}

// TitleType embedded fb2 type, represents title type fields
type TitleType struct {
	P         []string `xml:"p"`
	EmptyLine []string `xml:"empty-line"`
}

// PType embedded fb2 type, represents paragraph
type PType struct {
}

// ProtoFB2 type
type ProtoFB2 proto.PFB2
