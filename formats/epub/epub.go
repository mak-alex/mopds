package epub

import (
	// "archive/zip"
	// "errors"
	"fmt"
	// "image"
	// "image/gif"
	// "image/jpeg"
	// "image/png"
	// "os"
	// "path/filepath"
	// "strconv"
	// "time"

	// "golang.org/x/tools/godoc/vfs/zipfs"
	//
	"bitbucket.org/enlab/mopds/formats"
	"bitbucket.org/enlab/mopds/models"
	// "github.com/beevik/etree"
)

func indexer(filename string) (book *models.Book, err error) {
	defer func() {
		if r := recover(); r != nil {
			book = nil
			err = fmt.Errorf("Unknown error parsing book. Skipping. Error: %s", r)
		}
	}()

	// var title string
	// var author string
	// var seriesName string
	// var seriesIndex float64
	// var publisher string
	// var description string
	// var hasCover bool
	// var modtime time.Time

	// if file, err := os.Stat(filename); err == nil {
	// modtime = file.ModTime()
	// }
	//
	// zr, err := zip.OpenReader(filename)
	// if err != nil {
	// return nil, err
	// }
	//
	// zfs := zipfs.New(zr, "epub")
	//
	// rsk, err := zfs.Open("/META-INF/container.xml")
	// if err != nil {
	// return nil, err
	// }
	// defer rsk.Close()
	//
	// container := etree.NewDocument()
	// _, err = container.ReadFrom(rsk)
	// if err != nil {
	// return nil, err
	// }
	//
	// rootfile := ""
	// for _, e := range container.FindElements("//rootfiles/rootfile[@full-path]") {
	// rootfile = e.SelectAttrValue("full-path", "")
	// }
	//
	// if rootfile == "" {
	// return nil, errors.New("Cannot parse container")
	// }
	//
	// rrsk, err := zfs.Open("/" + rootfile)
	// if err != nil {
	// return nil, err
	// }
	// defer rrsk.Close()
	//
	// opfdir := filepath.Dir(rootfile)
	//
	// opf := etree.NewDocument()
	// _, err = opf.ReadFrom(rrsk)
	// if err != nil {
	// return nil, err
	// }
	//
	// title = filepath.Base(filename)
	// for _, e := range opf.FindElements("//title") {
	// title = e.Text()
	// break
	// }
	// for _, e := range opf.FindElements("//creator") {
	// author = e.Text()
	// break
	// }
	// for _, e := range opf.FindElements("//publisher") {
	// publisher = e.Text()
	// break
	// }
	// for _, e := range opf.FindElements("//description") {
	// description = e.Text()
	// break
	// }
	// for _, e := range opf.FindElements("//meta[@name='calibre:series']") {
	// seriesName = e.SelectAttrValue("content", "")
	// break
	// }
	// for _, e := range opf.FindElements("//meta[@name='calibre:series_index']") {
	// i, err := strconv.ParseFloat(e.SelectAttrValue("content", "0"), 64)
	// if err == nil {
	// seriesIndex = i
	// break
	// }
	// }

	/*for _, e := range opf.FindElements("//meta[@name='cover']") {
		coverid := e.SelectAttrValue("content", "")
		if coverid != "" {
			for _, f := range opf.FindElements("//[@id='" + coverid + "']") {
				coverPath := f.SelectAttrValue("href", "")
				if coverPath != "" {
					cr, err := zfs.Open("/" + opfdir + "/" + coverPath)
					if err != nil {
						continue
					}
					defer cr.Close()

					ext := filepath.Ext(coverPath)
					if ext == ".jpeg" {
						ext = ".jpg"
					}

					switch ext {
					case ".jpg":
						coverTmp, err = jpeg.Decode(cr)
						if err != nil {
							continue
						}
					case ".gif":
						coverTmp, err = gif.Decode(cr)
						if err != nil {
							continue
						}
					case ".png":
						coverTmp, err = png.Decode(cr)
						if err != nil {
							continue
						}
					}

					hasCover = true
				}
			}
			break
		}
	}*/
	catalog_id := uint(0)
	path := ""
	filetype := ""
	docdate := ""
	lang := ""
	filesize := "0"
	lib_id := "0"
	catalog_type := uint(0)
	ser_no := "0"
	del := ""
	update := false
	title := "test"
	return models.NewBook(catalog_id, filename, path, filetype, title, docdate, lang, filesize, lib_id, catalog_type, ser_no, del, update), nil
}

func init() {
	formats.RegisterFormat(&formats.Format{
		Glob:      "**/*.epub",
		Extension: ".epub",
		Indexer:   indexer,
	})
}
