package datastore

import (
	"fmt"
	_ "os"
	"strconv"
	"strings"

	"github.com/mak-alex/mopds/models"
	"github.com/mak-alex/mopds/modules/paginate"
	"github.com/mak-alex/mopds/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// TODO: вынести повторяющееся в отдельные методы и т.д.
const (
	DEFAULT_PAGE_SIZE int = 20
	MAX_PAGE_SIZE     int = 1000
	CAT_NORMAL        int = 0
	CAT_ZIP           int = 1
	CAT_INPX          int = 2
	CAT_INP           int = 3
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

	book.Catalog.CatType = 0
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

	/*annotations := []models.Annotation{}
	for _, annotation := range book.Annotations {
		filledAnnotation := models.Annotation{}
		store.db.FirstOrCreate(&filledAnnotation, annotation)
		annotations = append(annotations, filledAnnotation)
	}
	book.Annotations = annotations*/

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

func (store *dbStore) GetRandomBooksByAuthorID(authorID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bgenres on bgenres.book_id = books.id left join genres on bgenres.genre_id=genres.id")

	search = search.Where("authors.ID=?", authorID)

	//search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Order("RANDOM()").Find(&result)

	//p.Items = result
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}

	return p, nil
}

func (store *dbStore) GetRandomBooksByGenreID(genreID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bgenres on bgenres.book_id = books.id left join genres on bgenres.genre_id=genres.id")

	search = search.Where("genres.ID=?", genreID)

	//search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Order("RANDOM()").Find(&result)

	//p.Items = result
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}

	return p, nil
}

func (store *dbStore) GetRandomBooksBySerieID(serieID uint, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")

	search = search.Where("series.ID=?", serieID)

	//search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Order("RANDOM()").Find(&result)

	//p.Items = result
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}

	return p, nil
}

func (store *dbStore) GetBooksBySerie(title string, series string, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")

	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("LOWER(title) LIKE ?", "%"+term+"%")
	}
	for _, term := range utils.SplitBySeparators(strings.ToLower(series)) {
		search = search.Where("LOWER(ser) LIKE ?", "%"+term+"%")
	}

	//search = addParams(search, params).Group("books.id")

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}
	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Preload("Catalog").Order("ser_no, title").Find(&result)

	//p.Items = result
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}

	return p, nil
}

func (store *dbStore) GetBooksByLibID(libID string, noDetails bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books")
	search = search.Where("lib_id = ?", libID)

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}

	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Find(&result)
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}
	return p, nil
}

func (store *dbStore) GetAuthors(author string, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Author{}
	search := store.db.Select("authors.*").Table("authors")
	for _, term := range utils.SplitBySeparators(strings.ToLower(author)) {
		search = search.Where("LOWER(full_name) LIKE ?", "%"+term+"%")
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

	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Order("full_name")
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

func (store *dbStore) GetSeries(serie string, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Serie{}

	search := store.db.Select("series*").Table("series")
	for _, term := range utils.SplitBySeparators(strings.ToLower(serie)) {
		search = search.Where("LOWER(ser) LIKE ?", "%"+term+"%")
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
	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Select("series.*").Table("series").Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) GetSummary() (models.Summary, error) {
	summary := models.Summary{}

	authors_count := 0
	books_count := 0
	genres_count := 0
	series_count := 0
	catalogs_count := 0

	store.db.Table("authors").Count(&authors_count)
	store.db.Table("books").Count(&books_count)
	store.db.Table("genres").Count(&genres_count)
	store.db.Table("series").Count(&series_count)
	store.db.Table("catalogs").Count(&catalogs_count)

	summary.Authors = authors_count
	summary.Books = books_count
	summary.Genres = genres_count
	summary.Series = series_count
	summary.Catalogs = catalogs_count

	return summary, nil
}

func (store *dbStore) GetGenresMenu() ([]models.ItemMenu, error) {
	item := []models.ItemMenu{}
	found := make(map[string]bool)

	result := []models.Genre{}
	store.db.Select("genres.*").Table("genres").Find(&result)
	// create first level menu
	for i, genre := range result {
		i++
		if !found[genre.Section] {
			item = append(item, models.ItemMenu{ID: strconv.Itoa(i), Value: genre.Section})
			found[genre.Section] = true
		}
	}
	// create sublevel menu
	for j, genre := range result {
		j++
		for i, _ := range item {
			if genre.Section == item[i].Value {
				item[i].Data = append(item[i].Data, models.SubItemMenu{ID: fmt.Sprintf("%s.%s", strconv.Itoa(j), strconv.Itoa(i)), GenreID: genre.ID, Value: genre.Subsection})
			}
		}
	}
	return item, nil
}

func (store *dbStore) GetGenres(genre string, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Genre{}
	search := store.db.Select("genres.*").Table("genres")
	for _, term := range utils.SplitBySeparators(strings.ToLower(genre)) {
		search = search.Where("LOWER(genre) LIKE ? OR LOWER(section) LIKE ? OR LOWER(subsection) LIKE ?", "%"+term+"%", "%"+term+"%", "%"+term+"%")
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

	search.Count(&count)

	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	search.Select("genres.*").Table("genres").Find(&result)
	p.Items = result
	return p, nil
}

func (store *dbStore) GetBooks(title string, noDetails bool, random bool, page int, per_page int) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}

	search := store.db.Select("books.*").Table("books").
		// select * from books left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("LOWER(title) LIKE ?", "%"+term+"%")
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
	for _, term := range utils.SplitBySeparators(strings.ToLower(title)) {
		search = search.Where("LOWER(title) LIKE ?", "%"+term+"%")
	}

	search.Count(&count)
	p := paginate.NewPaginatedList(page, per_page, count)
	search = search.Limit(p.Limit())
	search = search.Offset(p.Offset())
	if random {
		search.Group("books.id").Select("books.*").Table("books").Order("RANDOM()").Find(&result)
	} else {
		search.Group("books.id").Select("books.*").Table("books").Find(&result)
	}

	// utils.PrintJson(result, true)
	if !noDetails {
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}
	return p, nil
}

func (store *dbStore) GetBooksByGenreID(genreID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bgenres on bgenres.book_id = books.id left join genres on bgenres.genre_id=genres.id")
	search = search.Where("genres.ID=?", genreID)
	search.Count(&count)

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
	if random {
		search.Preload("Catalog").Order("RANDOM()").Find(&result)
	} else {
		search.Preload("Catalog").Find(&result)
	}

	if !noDetails {
		//p.Items = store.fillBooksDetails(result, false, false, false, false)
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}
	p.Items = result
	return p, nil
}

func (store *dbStore) GetBooksBySerieID(serieID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	search = search.Where("series.ID=?", serieID)
	search.Count(&count)

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
	if random {
		search.Preload("Catalog").Order("RANDOM()").Find(&result)
	} else {
		search.Preload("Catalog").Find(&result)
	}

	if !noDetails {
		//p.Items = store.fillBooksDetails(result, false, false, false, false)
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
	}
	p.Items = result
	return p, nil
}

func (store *dbStore) GetBooksByAuthorID(authorID uint, noDetails bool, random bool, page int, per_page int, params models.Search) (*paginate.PaginatedList, error) {
	count := 0
	result := []models.Book{}
	search := store.db.Select("books.*").Table("books").
		Joins("left join bauthors on books.id=bauthors.book_id left join authors on authors.id=bauthors.author_id left join bcovers on books.id=bcovers.book_id left join covers on covers.id=bcovers.book_id left join bseries on bseries.book_id = books.id left join series on bseries.serie_id=series.id")
	search = search.Where("authors.ID=?", authorID)
	search.Count(&count)

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
	if random {
		search.Preload("Catalog").Order("RANDOM()").Find(&result)
	} else {
		search.Preload("Catalog").Find(&result)
	}

	if !noDetails {
		//p.Items = store.fillBooksDetails(result, false, false, false, false)
		p.Items = store.fillBooksDetails(result, true, true, true, true)
	} else {
		p.Items = store.fillBooksDetails(result, false, false, false, false)
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
		Where("lib_id = ? and file_name = ?", book.LibID, book.Catalog.CatName).
		First(found)
	book.ID = found.ID
	book.CatalogID = found.CatalogID
	book.Catalog = models.Catalog{}
	// if found != book {
	// store.db.Save(book)
	// }
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

func (store *dbStore) IsFileExist(fileName string) bool {
	contObj := new(models.Book)
	store.db.Where("file_name = ?", fileName).First(&contObj)
	return contObj.ID > 0
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
