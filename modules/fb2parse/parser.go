package fb2parse

import (
  "os"
  "fmt"
  "bytes"
  "strings"
  "net/http"
  "encoding/xml"
  xs "github.com/huandu/xstrings"
  "golang.org/x/net/html/charset"
  "bitbucket.org/enlab/mopds/utils"
  "bitbucket.org/enlab/mopds/models"
)

type BookInfo struct {
  FirstName string
  LastName  string
  Authors   []models.Author
  Title     string
  Sequence  string
  SerNo     string
  DocDate   string
  Language  string
  Genres    utils.Genres
  Annotation string
  Covers     []models.Cover
  FileSize  int
  CoverID   string
  CoverContentType string
}

/*
IsZipFile checks if the file is ZIP archive.
Returns true is the file is ZIP or GZIP archive and false otherwise
*/
func IsZipFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
	}

	fileType := http.DetectContentType(buff)
	switch fileType {
	case "application/x-gzip", "application/zip":
		return true
	default:
		return false
	}
}

func isInBookInfo(path []string) bool {
	if len(path) < 3 {
		return false
	}

	return path[0] == "FictionBook" &&
		path[1] == "description" &&
    path[2] == "title-info"
}

func isInBookContent(path []string) bool {
	if len(path) < 2 {
		return false
	}

	return path[0] == "FictionBook" &&
		path[1] == "body"
}

func isInside(path []string, sectionName string) bool {
	n := len(path) - 1
	if n < 0 {
		return false
	}

	for n >= 0 {
		if path[n] == sectionName {
			return true
		}
		if path[n] == "p" || path[n] == "emphasis" || path[n] == "binary" || path[n] == "text-author" || path[n] == "strong" {
			n--
		} else {
			break
		}
	}

	return false
}

func ParseBook(file []byte, genres utils.Genres, parseBody bool) (BookInfo, []string) {
	//isZip := IsZipFile(fileName)

	lines := make([]string, 0)
	var binfo BookInfo
	tags := make([]string, 0, 10)

	var decoder *xml.Decoder
	reader := bytes.NewReader(file)
  decoder = xml.NewDecoder(reader)
  decoder.CharsetReader = charset.NewReaderLabel

	var currLine string
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			if !parseBody && se.Name.Local == "body" {
				return binfo, lines
			}

			if se.Name.Local == "empty-line" {
				lines = append(lines, "")
				currLine = ""
			} else if se.Name.Local == "section" {
				lines = append(lines, "{{section}}")
				currLine = ""
			} else if se.Name.Local == "emphasis" || se.Name.Local == "strong" {
				currLine += "{{emon}}"
			} else if se.Name.Local == "date" {
				for i := 0; i < len(se.Attr); i++ {
					if se.Attr[i].Name.Local == "value" {
						binfo.DocDate = se.Attr[i].Value
					}
				}
			} else if se.Name.Local == "sequence" {
				for i := 0; i < len(se.Attr); i++ {
					if se.Attr[i].Name.Local == "name" {
						binfo.Sequence = se.Attr[i].Value
					}
          if se.Attr[i].Name.Local == "number" {
            binfo.SerNo = se.Attr[i].Value
          }
				}
			} else if se.Name.Local == "binary" {
				for i := 0; i < len(se.Attr); i++ {
          if se.Attr[i].Name.Local == "id" {
            binfo.CoverID = se.Attr[i].Value
          }
          if se.Attr[i].Name.Local == "content-type" {
            binfo.CoverContentType = se.Attr[i].Value
          }
				}
			} else {
				if se.Name.Local == "text-author" && isInside(tags, "epigraph") {
					currLine = "{{epiauth}}"
				} else if se.Name.Local == "p" {
					if isInside(tags, "epigraph") {
						currLine = "{{epi}}"
					} else if isInside(tags, "title") {
						currLine = "{{title}}"
					} else if isInside(tags, "binary") {
						currLine = "{{binary}}"
					} else {
						currLine = ""
					}
				} else {
					currLine = ""
			 	}
			}
			tags = append(tags, se.Name.Local)
		case xml.EndElement:
			if tags[len(tags)-1] != se.Name.Local {
				panic("Invalid fb2")
			}
			tags = tags[:len(tags)-1]
			if isInBookInfo(tags) {
				if se.Name.Local == "genre" {
					binfo.Genres = append(binfo.Genres, genres.GetDetail(currLine))
				} else if se.Name.Local == "first-name" && isInside(tags, "author") {
					binfo.FirstName = currLine
				} else if se.Name.Local == "last-name" && isInside(tags, "author") {
          a := fmt.Sprintf("%s %s", binfo.FirstName, currLine)
          binfo.Authors = append(binfo.Authors, models.Author{FullName: a, SearchFullName: strings.ToUpper(a) })
				}
        if se.Name.Local == "book-title" {
					binfo.Title = currLine
				} else if se.Name.Local == "lang" {
					binfo.Language = currLine
				} else if se.Name.Local == "annotation" {
					binfo.Annotation = currLine
        }
			} else if se.Name.Local == "binary" {
        if len(binfo.Covers) == 0 {
          binfo.Covers = append(binfo.Covers, models.Cover{Name: binfo.CoverID, Value: currLine, ContentType: binfo.CoverContentType})
        }
      } else if isInBookContent(tags) {
				if se.Name.Local == "body" {
					return binfo, lines
				} else if se.Name.Local == "emphasis" || se.Name.Local == "strong" {
					currLine += "{{emoff}}"
				} else {
					if currLine != "" {
						lines = append(lines, currLine)
					}
					currLine = ""
				}
			} else {
				currLine = ""
			}
		case xml.CharData:
			ss := string(se)
			newLines := xs.Count(ss, "\n\r ")
			if newLines != len(ss) {
				ss = xs.Squeeze(xs.Translate(ss, "\n\r", "  "), " ")
				currLine += ss
			}
		}
	}

	return binfo, lines
}
