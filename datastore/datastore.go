package datastore

import (
	"fmt"
	_ "os"
	"strings"

	"bitbucket.org/enlab/mopds/paginate"
	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// TODO: вынести повторяющееся в отдельные методы и т.д.
const (
	DEFAULT_PAGE_SIZE int = 25
	MAX_PAGE_SIZE     int = 1000
)

type dbStore struct {
	db *gorm.DB
}

func addParams(search *gorm.DB, params models.Search) *gorm.DB {
	if !params.Deleted {
		search = search.Where("del='0'")
	}

	if len(params.Langs) > 0 {
		for i, _ := range params.Langs {
			params.Langs[i] = "'" + params.Langs[i] + "'"
		}
		search = search.Where("lang in (" + strings.Join(params.Langs, ",") + ")")
	}
	return search
}

func (store *dbStore) PutBook(book *models.Book) (err error) {
	tx := store.db.Begin()

	store.db.FirstOrCreate(&book.Catalog, book.Catalog)
	authors := []models.Author{}
	for _, author := range book.Authors {
		filledAuthor := models.Author{}
		store.db.FirstOrCreate(&filledAuthor, author)
		authors = append(authors, filledAuthor)
	}
	book.Authors = authors

	genres := []models.Genre{}
	for _, genre := range book.Genres {
		filledGenre := models.Genre{}
		store.db.FirstOrCreate(&filledGenre, genre)
		genres = append(genres, filledGenre)
	}
	book.Genres = genres

	covers := []models.Cover{}
	for _, cover := range book.Covers {
		filledCover := models.Cover{}
		store.db.FirstOrCreate(&filledCover, cover)
		covers = append(covers, filledCover)
	}
	book.Covers = covers

	store.db.Create(&book)

	tx.Commit()

	return err
}

func (store *dbStore) fillBookDetails(book *models.Book, fillGenres bool, fillCovers bool, fillCatalogs bool, fillAnnotations bool) {
	store.db.Select("authors.*").Model(book).Related(&book.Authors, "Authors")
	for j, a := range book.Authors {
		book.Authors[j].FullName = utils.UpperInitialAll(a.FullName)
	}
	if fillGenres {
		store.db.Select("genres.*").Model(book).Related(&book.Genres, "Genres")
	}
	if fillCovers {
		store.db.Select("covers.*").Model(book).Related(&book.Covers, "Covers")
	}
	if fillCatalogs {
		store.db.Select("catalogs.*").Model(book).Related(&book.Catalog, "Catalog")
	}
	if fillAnnotations {
		store.db.Select("annotations.*").Model(book).Related(&book.Annotations, "Annotations")
	}
}

func (store *dbStore) fillBooksDetails(books []models.Book, fillGenres bool, fillCovers bool, fillCatalogs bool, fillAnnotations bool) []models.Book {
	for i, _ := range books {
		store.fillBookDetails(&books[i], fillGenres, fillCovers, fillCatalogs, fillAnnotations)
	}

	return books
}

func (store *dbStore) FindBooks(params models.Search) ([]models.Book, error) {
	title := params.Title
	authors := params.Author
	limit := params.Limit

	// utils.PrintJson(params)

	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		// select * from books left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("title LIKE ?", "%"+term+"%")
	}
	for _, term := range utils.SplitBySeparators(strings.ToLower(authors)) {
		search = search.Where("full_name LIKE ?", "%"+term+"%")
	}

	search = addParams(search, params).Group("books.id")

	if limit > 0 {
		search = search.Limit(limit)
	}
	search.Preload("Catalog").Order("title").Find(&result)

	result = store.fillBooksDetails(result, false, false, false, false)
	return result, nil
}

func (store *dbStore) FindBooksSeries(params models.Search) ([]models.Book, error) {
	title := params.Title
	series := params.Series
	limit := params.Limit

	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("title LIKE ?", "%"+term+"%")
	}
	for _, term := range utils.SplitBySeparators(strings.ToLower(series)) {
		search = search.Where("LOWER(ser) LIKE ?", "%"+term+"%")
	}

	search = addParams(search, params).Group("books.id")

	if limit > 0 {
		search = search.Limit(limit)
	}
	search.Preload("Catalog").Order("ser_no, title").Find(&result)

	return result, nil
}

func (store *dbStore) FindBooksByLibID(libID string) ([]models.Book, error) {
	result := []models.Book{}
	store.db.Select("books.*").Table("books").
		Where("lib_id = ?", libID).
		Find(&result)
	result = store.fillBooksDetails(result, true, true, true, true)
	return result, nil
}

func (store *dbStore) GetAuthors(author string, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Author{}
	search := store.db.Order("full_name")
	for _, term := range utils.SplitBySeparators(strings.ToLower(author)) {
		search = search.Where("full_name LIKE ?", "%"+term+"%")
	}
	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Find(&result).Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) GetCover(coverID uint) (*models.Cover, error) {
	result := new(models.Cover)
	store.db.First(result, coverID)
	if result.ID > 0 {
		//result.FullName = utils.UpperInitialAll(result.FullName)
		return result, nil
	}
	return nil, fmt.Errorf("No cover found")
}

func (store *dbStore) GetGenre(genreID uint) (*models.Genre, error) {
	result := new(models.Genre)
	store.db.First(result, genreID)
	if result.ID > 0 {
		result.Genre = utils.UpperInitialAll(result.Genre)
		return result, nil
	}
	return nil, fmt.Errorf("No genre found")
}

func (store *dbStore) GetSerie(serieID uint) (*models.Serie, error) {
	result := new(models.Serie)
	store.db.First(result, serieID)
	if result.ID > 0 {
		result.Ser = utils.UpperInitialAll(result.Ser)
		return result, nil
	}
	return nil, fmt.Errorf("No serie found")
}

func (store *dbStore) GetAuthor(authorID uint) (*models.Author, error) {
	result := new(models.Author)
	store.db.First(result, authorID)
	if result.ID > 0 {
		result.FullName = utils.UpperInitialAll(result.FullName)
		return result, nil
	}
	return nil, fmt.Errorf("No author found")
}

func (store *dbStore) GetSeries(page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Serie{}

	search := store.db.Order("ser")
	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Find(&result).Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Select("series.*").Table("series").Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) GetGenres(page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0;
	result := []models.Genre{}

	search := store.db.Order("section")
	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Find(&result).Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Select("genres.*").Table("genres").Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) GetBooks(title string, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0;
	result := []models.Book{}

	search := store.db.Order("title")
	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("title LIKE ?", "%"+term+"%")
	}

	search.Find(&result).Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Select("books.*").Table("books").Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) ListGenreBooks(genreID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bgenres on bgenres.book_id = books.id left join genres on bgenres.genre_id=genres.id")
	search = search.Where("genres.ID=?", genreID)
	search.Find(&result).Count(&count)

	search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page >= count {
		per_page = count
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Find(&result)

	if !noDetails {
		result = store.fillBooksDetails(result, false, false, false, false)
	}
	p.Items = result
	return p, nil
}

func (store *dbStore) ListSerieBooks(serieID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	search = search.Where("series.ID=?", serieID)
	search.Find(&result).Count(&count)

	search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page >= count {
		per_page = count
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Find(&result)

	if !noDetails {
		result = store.fillBooksDetails(result, false, false, false, false)
	}
	p.Items = result
	return p, nil
}

func (store *dbStore) ListAuthorBooks(authorID uint, noDetails bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	search = search.Where("authors.ID=?", authorID)
	search.Find(&result).Count(&count)

	search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page >= count {
		per_page = count
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Find(&result)

	if !noDetails {
		result = store.fillBooksDetails(result, false, false, false, false)
	}
	p.Items = result
	return p, nil
}

func (store *dbStore) GetBook(bookID uint) (*models.Book, error) {
	result := new(models.Book)
	store.db.Select("*").Preload("Catalog").First(result, bookID)
	store.fillBookDetails(result, true, true, true, true)
	if result.ID > 0 {
		return result, nil
	}
	return nil, fmt.Errorf("No book found")
}

func (store *dbStore) UpdateBook(book *models.Book) (*models.Book, error) {
	found := new(models.Book)
	store.db.Select("distinct books.*").Table("books").
		Joins("left join catalogs on catalogs.id = books.catalog_id").
		Where("lib_id = ? and filename = ?", book.LibID, book.Catalog.CatName).
		First(found)
	book.ID = found.ID
	book.CatalogID = found.CatalogID
	book.Catalog = models.Catalog{}
	if found != book {
		store.db.Save(book)
	}
	return book, nil
}

func (store *dbStore) GetLangs() ([]string, error) {
	var result []string
	found := []models.Book{}
	store.db.Select("distinct books.lang").
		Table("books").Where("lang <> ''").
		Order("lang").
		Find(&found)

	for _, book := range found {
		result = append(result, book.Lang)
	}
	return result, nil
}

func (store *dbStore) IsCatalogExist(fileName string) bool {
	contObj := new(models.Catalog)
	store.db.Where("cat_name = ?", fileName).First(&contObj)
	return contObj.ID > 0
}

func (store *dbStore) Close() {
	store.db.Close()
}

// NewDBStore creates new instance of datastorer
func NewDBStore(config *models.DBConfig) (DataStorer, error) {
	fmt.Println(config.DBType, config.DBParams)
	db, err := gorm.Open(config.DBType, config.DBParams)
	if err == nil {
		db.DB()
		db.AutoMigrate(&models.Author{}, &models.Book{}, &models.Annotation{}, &models.Cover{}, &models.Catalog{}, &models.Genre{}, &models.Serie{})
		db.LogMode(config.DBLog)
	}
	result := new(dbStore)
	result.db = db

	return result, err
}
