package formats

import (
	// "image"

	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/utils"
)

// IndexerFunc is a function which takes a filename and returns a Book and a Cover, if it is present.
type IndexerFunc func(filename string, filesize string, filetype string, path string, file []byte, genres utils.Genres, update bool) (book *models.Book, err error)

// Format represents a handler for an ebook format.
type Format struct {
	Glob      string
	Extension string // Needs to include the leading period. Also used as the unique identifier.
	Indexer   IndexerFunc
}

// FormatList represents a list of formats.
type FormatList []*Format

// Formats is the list of registered formats.
var Formats FormatList

// RegisterFormat adds a format to the Formats list.
// This function does not do anything if there is already a format with the given extension.
func RegisterFormat(f *Format) {
	for _, i := range Formats {
		if i.Extension == f.Extension {
			return
		}
	}

	Formats = append(Formats, f)
}
