package datastore

import (
	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/modules/paginate"
)

// DataStorer interface for data layer
type DataStorer interface {
	// One
	GetAuthor(authorID uint) (*models.Author, error)
	GetBook(bookID uint) (*models.Book, error)
	GetSerie(serieID uint) (*models.Serie, error)
	GetGenre(genreID uint) (*models.Genre, error)
	// List
	GetAuthors(author string, page int, per_page int) (*paginate.PaginatedList, error)
	GetBooks(title string, noDetails bool, random bool, page int, per_page int) (*paginate.PaginatedList, error)
	GetGenres(genre string, page int, per_page int) (*paginate.PaginatedList, error)
	GetSeries(serie string, page int, per_page int) (*paginate.PaginatedList, error)
	GetLangs() ([]string, error)
	GetSummary() (models.Summary, error)
	// Exists?
	IsCatalogExist(fileName string) bool
	IsFileExist(fileName string) bool
	// List by...
	GetBooksByLibID(libID string, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error)
	GetBooksByAuthorID(authorID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	GetBooksBySerieID(authorID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	GetBooksByGenreID(genreID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	GetBooksBySerie(title string, series string, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error)
	GetRandomBooksByAuthorID(authorID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error)
	GetRandomBooksByGenreID(genreID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error)
	GetRandomBooksBySerieID(serieID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error)
	// Add/Update/Delete
	PutBook(*models.Book) error
	UpdateBook(*models.Book) (*models.Book, error)
	// Menu
	GetGenresMenu() ([]models.ItemMenu, error)
	// Others
	Close()
}
