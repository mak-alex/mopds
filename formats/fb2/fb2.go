package epub

import (
	// "archive/zip"
	"errors"
	// "io"
	"fmt"
	// "image"
	// "image/gif"
	// "image/jpeg"
	// "image/png"
	// "os"
	// "path/filepath"
	// "strconv"
	"strings"
	// "time"
	"golang.org/x/net/html/charset"

	// "golang.org/x/tools/godoc/vfs/zipfs"
	//
	"bitbucket.org/enlab/mopds/formats"
	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/utils"
	"github.com/beevik/etree"
)

func indexer(filename string, filesize string, filetype string, path string, file []byte, genres utils.Genres, update bool) (book *models.Book, err error) {
	defer func() {
		if r := recover(); r != nil {
			book = nil
			err = fmt.Errorf("Unknown error parsing book. Skipping. Error: %s", r)
		}
	}()

	var title string
	// var author string
	var seriesName string
	var ser_no string
	var docdate string
	var description string
	var lang string
	// var modtime time.Time
	var Authors []models.Author
	var Annotations []models.Annotation
	var Covers []models.Cover
	var Genres []models.Genre
	var Series []models.Serie

	// if file, err := os.Stat(filename); err == nil {
	// modtime = file.ModTime()
	// }

	container := etree.NewDocument()
	container.ReadSettings.CharsetReader = charset.NewReaderLabel
	err = container.ReadFromBytes(file)
	if err != nil {
		return nil, err
	}
	//rootfile := container.SelectElement("FictionBook")
	rootfile := ""
	for _, e := range container.FindElements("//FictionBook") {
		rootfile = e.SelectAttrValue("xmlns", "")
	}

	if rootfile == "" {
		return nil, errors.New("Cannot parse container")
	}

	for _, e := range container.FindElements("//book-title") {
		title = e.Text()
		break
	}
	for _, e := range container.FindElements("//title-info/genre") {
		genre := e.Text()
		g := genres.GetDetail(genre)
		Genres = append(Genres, models.Genre{Genre: g.Genre, Section: g.Section, Subsection: g.Subsection})
	}
	for _, e := range container.FindElements("//title-info/author") {
		f := e.SelectElements("first-name")
		firstname := ""
		lastname := ""
		for _, fn := range f {
			firstname = fn.Text()
		}
		l := e.SelectElements("last-name")
		for _, ln := range l {
			lastname = ln.Text()
		}
		author := fmt.Sprintf("%s %s", firstname, lastname)
		Authors = append(Authors, models.Author{FullName: strings.ToLower(author), SearchFullName: strings.ToUpper(author)})
	}
	for _, e := range container.FindElements("//title-info/annotation/*") {
		description = e.Text()
		Annotations = append(Annotations, models.Annotation{Value: description})
		break
	}
	for _, e := range container.FindElements("//title-info/sequence") {
		ser_no = e.SelectAttrValue("number", "")
		seriesName = e.SelectAttrValue("name", "")
		Series = append(Series, models.Serie{Ser: seriesName, SearchSer: strings.ToUpper(seriesName)})
		break
	}

	for _, e := range container.FindElements("//binary") {
		coverid := e.SelectAttrValue("id", "")
		covertype := e.SelectAttrValue("content-type", "")
		coverimage := e.Text()
		if strings.Index(coverid, "cover") == 0 && len(Covers) <= 0 {
			Covers = append(Covers, models.Cover{Name: coverid, ContentType: covertype, Value: coverimage})
		}
	}

	for _, e := range container.FindElements("//title-info/lang") {
		lang = e.Text()
	}
	for _, e := range container.FindElements("//title-info/date") {
		docdate = e.Text()
	}

	catalog_id := uint(0)
	lib_id := "0"
	catalog_type := uint(0)
	del := ""

	return models.NewBook(catalog_id, filename, path, filetype, title,
		docdate, lang, filesize, lib_id, catalog_type, ser_no, del, update, Authors, Genres, Series, Annotations, Covers), nil
}

func init() {
	formats.RegisterFormat(&formats.Format{
		Glob:      "**/*.fb2",
		Extension: ".fb2",
		Indexer:   indexer,
	})
}
