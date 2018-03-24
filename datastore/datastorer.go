package datastore

import (
	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/paginate"
)

// DataStorer interface for data layer
type DataStorer interface {
	PutBook(*models.Book) error
	UpdateBook(*models.Book) (*models.Book, error)
	FindBooks(models.Search) ([]models.Book, error)
	FindBooksSeries(models.Search) ([]models.Book, error)
	FindBooksByLibID(libID string) ([]models.Book, error)
	ListAuthorBooks(authorID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	ListSerieBooks(authorID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	ListGenreBooks(genreID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error)
	GetBook(bookID uint) (*models.Book, error)
	GetAuthor(authorID uint) (*models.Author, error)
	GetSerie(serieID uint) (*models.Serie, error)
	GetGenre(genreID uint) (*models.Genre, error)
	GetAuthors(author string, page int, per_page int) (*paginate.PaginatedList, error)
	GetBooks(title string, page int, per_page int) (*paginate.PaginatedList, error)
	GetGenres(page int, per_page int) (*paginate.PaginatedList, error)
	GetSeries(page int, per_page int) (*paginate.PaginatedList, error)
	GetLangs() ([]string, error)
	IsCatalogExist(fileName string) bool
	Close()
}
