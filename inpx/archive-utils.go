package inpx

import (
	"archive/zip"
	"fmt"
	"bitbucket.org/enlab/mopds/models"
	"io"
	"log"
	"os"
	"path/filepath"
)

func UnzipBookToWriter(dataDir string, book *models.Book, writer io.Writer) (err error) {
	container := book.Catalog.CatName
	fileName := book.FileName + "." + book.Format

	r, err := zip.OpenReader(filepath.Join(dataDir, container))
	if err != nil {
		log.Printf("Failed to open container %s\n", container)
		return fmt.Errorf("Failed to open container %s", container)
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
