package books

import (
	"archive/zip"
	"github.com/mak-alex/mopds/models"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// TODO: доработать!!!
func UnzipBookToWriter(dataDir string, book *models.Book, writer io.Writer) (err error) {
	container := book.Catalog.CatName
	catalog_path := book.Catalog.Path
	cat_ext := filepath.Ext(container)
	fileName := book.FileName + book.Format

	fmt.Println(cat_ext)
	if cat_ext == ".zip" {
		r, err := zip.OpenReader(filepath.Join(catalog_path, container))
		if err != nil {
			log.Printf("Failed to open container %s/%s\n", catalog_path, fileName)
			return fmt.Errorf("Failed to open container %s/%s", catalog_path, fileName)
		}
		defer r.Close()
		for _, file := range r.File {
			if file.FileInfo().Name() == fileName {
				rc, err := file.Open()
				if err != nil {
					return err
				}
				defer rc.Close()

				_, err = io.Copy(writer, rc)

				break
			}
		}
	} else if cat_ext == ".fb2" {
		rc, err := os.Open(filepath.Join(catalog_path, container))
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(writer, rc)
	}
	return err
}

func UnzipBookFile(dataDir string, book *models.Book, targetFolder string, rename bool) (err error) {
	var outName string
	if rename {
		authors := ""
		for _, a := range book.Authors {
			authors = authors + a.FullName
		}
		outName = authors + " - "
		if book.SerNo != "" {
			if len(book.SerNo) == 1 {
				book.SerNo = "0" + book.SerNo
			}

			outName = outName + "[" + book.SerNo + "] "
		}
		outName = outName + book.Title + "." + book.Format
	} else {
		outName = book.FileName + "." + book.Format
	}

	path := filepath.Join(targetFolder, outName)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	err = UnzipBookToWriter(dataDir, book, f)

	return err

}
