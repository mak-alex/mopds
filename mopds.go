package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mak-alex/mopds/models"
	"github.com/mak-alex/mopds/modules/books"
	"github.com/mak-alex/mopds/modules/datastore"
	"github.com/mak-alex/mopds/modules/new_rest"
	"github.com/mak-alex/mopds/utils"
	"github.com/namsral/flag"
)

// TODO: пофиксить проблему с кодировками при индексировании файло из архивов
// TODO: пофиксить проблему с дубликатами при индексировании (и при повтороном)
// TODO: добавить возможность индексировать без inpx
// TODO: добавить поддержку форматов epub/pdf
// TODO: добавить возможность индексировать директорию с *.{fb2,epub}
func findINPX(catalog string) []string {
	inpx_files, err := filepath.Glob(filepath.Join(catalog, "*.inpx"))
	if err != nil {
		log.Fatal(err)
	}
	return inpx_files
}

func setDatabaseConfig(Catalog string, username string, password string, host string, dbname string, dbtype string, dblog bool, sslmode string) *models.DBConfig {
	result := new(models.DBConfig)

	result.DBType = dbtype
	result.DBLog = dblog
	if dbtype == "sqlite3" {
		fileData, err := ioutil.ReadFile(filepath.Join(Catalog, "mopds_db.sqlite"))
		if err == nil {
			err = json.Unmarshal(fileData, result)
		}

		if err != nil { // fallback to sqlite
			result.DBParams = filepath.Join(Catalog, "mopds_db.sqlite")
		}
	} else if dbtype == "postgres" {
		if sslmode == "" {
			sslmode = "disable"
		}
		result.DBParams = fmt.Sprintf("user=%s password=%s DB.name=%s sslmode=%s", username, password, dbname, sslmode)
	}

	return result
}

func main() {
	curDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	var (
		config            string
		Catalog           string
		Page              int
		PerPage           int
		GetAuthor         uint
		GetBook           uint
		GetGenre          uint
		GetSerie          uint
		GetAuthors        bool
		GetBooks          bool
		GetGenres         bool
		GetSeries         bool
		GetBooksByAuthor  uint
		GetBooksByGenre   uint
		GetBooksBySerie   uint
		SearchAuthor      string
		SearchGenre       string
		SearchSerie       string
		Random            bool
		Stat              bool
		Listen            string
		Scan              bool
		Parse             bool
		Save              bool
		SearchLibID       string
		SearchBookByTitle string
		Verbose           bool
		Debug             bool
		About             bool
		Host              string
		Username          string
		Password          string
		DBName            string
		DBType            string
		DBLog             bool
		SSLMode           string
	)

	flag.StringVar(&config, "config", "./conf/mopds.conf", "Default configuration file")
	flag.StringVar(&Catalog, "catalog", "", "Directory of library (mandatory)")
	flag.IntVar(&Page, "page", 0, "Pagination 1...n")
	flag.IntVar(&PerPage, "per_page", 25, "Limit results (-1 for no limit)")
	flag.UintVar(&GetAuthor, "get_author", 0, "Get author by id")
	flag.UintVar(&GetBook, "get_book", 0, "Get book by id")
	flag.UintVar(&GetGenre, "get_genre", 0, "Get genre by id")
	flag.UintVar(&GetSerie, "get_serie", 0, "Get serie by id")
	flag.BoolVar(&GetAuthors, "get_authors", false, "List all authors")
	flag.BoolVar(&GetBooks, "get_books", false, "List all books")
	flag.BoolVar(&GetGenres, "get_genres", false, "List all genres")
	flag.BoolVar(&GetSeries, "get_series", false, "List all series")
	flag.UintVar(&GetBooksByAuthor, "get_books_by_author", 0, "List all author's books by id")
	flag.UintVar(&GetBooksByGenre, "get_books_by_genre", 0, "List all genre's books by id")
	flag.UintVar(&GetBooksBySerie, "get_books_by_serie", 0, "List all serie's books by id")
	flag.StringVar(&SearchAuthor, "search_author", "", "Search authors, or books by author if comes with search-title")
	flag.StringVar(&SearchGenre, "search_genre", "", "Search genre by genre name or section, or subsection")
	flag.StringVar(&SearchSerie, "search_serie", "", "Search serie by serie name")
	flag.BoolVar(&Random, "random", false, "Random for -get_books/-get_books_by_genre/-get_books_by_author/-get_books_by_serie)")
	flag.BoolVar(&Stat, "stat", false, "Book library statistics")
	flag.StringVar(&Listen, "listen", ":8000", "Set server listen address:port")
	flag.BoolVar(&Scan, "scan", false, "Start scanning books")
	flag.BoolVar(&Parse, "inpx", true, "Parse inpx to the local database. If not used, mopds continue work to zip archives and *.fb2 files")
	flag.BoolVar(&Save, "save", false, "Save book file to the disk (ex.: --get_book 1 --save)")
	flag.StringVar(&SearchLibID, "search_lib_id", "", "Search book(s) by its libId")
	flag.StringVar(&SearchBookByTitle, "search_book", "", "Search books by their title")
	flag.BoolVar(&Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&Debug, "debug", false, "Debug output")
	flag.BoolVar(&About, "about", false, "About author and this project")
	flag.StringVar(&Host, "host", "modps", "IP address for connect to database")
	flag.StringVar(&Username, "username", "mopds", "Username for connect to database")
	flag.StringVar(&Password, "password", "mopds", "Password for connect to database")
	flag.StringVar(&DBType, "dbtype", "sqlite3", "Type used database: sqlite3, mysql or postgres")
	flag.StringVar(&DBName, "database", "", "Database name connect to database")
	flag.StringVar(&SSLMode, "sslmode", "", "Whether to use ssl mode or not, here's the question: disable or enable")

	flag.Parse()

	DBLog = Debug
	conf := setDatabaseConfig(Catalog, Username, Password, Host, DBName, DBType, DBLog, SSLMode)
	store, err := datastore.NewDBStore(conf)
	if err != nil {
		log.Fatalln("Failed to open database")
	}
	defer store.Close()

	if Scan {
		log.Printf("Opening %s to parse data\n", Catalog)
		go func() {
			books.ProcessALL(Catalog, Parse, store)
		}()
		rest.NewRestService(Listen, store, Catalog).StartListen()
	} else if SearchLibID != "" {
		result, err := store.GetBooksByLibID(SearchLibID, true, Page, PerPage)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if SearchBookByTitle != "" {
		result, err := store.GetBooksBySerie(SearchBookByTitle, SearchAuthor, true, Page, PerPage)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if SearchAuthor != "" {
		result, err := store.GetAuthors(SearchAuthor, Page, PerPage)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetAuthors {
		result, err := store.GetAuthors("", Page, PerPage)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetBooks {
		result, err := store.GetBooks("", true, Random, Page, PerPage)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetGenres {
		result, err := store.GetGenres("", Page, PerPage) // фильтр добавить
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetSeries {
		result, err := store.GetSeries("", Page, PerPage) // фильтр добавить
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetBooksByAuthor != 0 {
		result, err := store.GetBooksByAuthorID(GetBooksByAuthor, true, Random, Page, PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetBooksByGenre != 0 {
		result, err := store.GetBooksByGenreID(GetBooksByGenre, true, Random, Page, PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetBooksBySerie != 0 {
		result, err := store.GetBooksBySerieID(GetBooksBySerie, true, Random, Page, PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetAuthor != 0 {
		result, err := store.GetAuthor(GetAuthor)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetBook != 0 {
		result, err := store.GetBook(GetBook)
		if err == nil {
			utils.PrintJson(result, true)
			if Save {
				err = books.UnzipBookFile(Catalog, result, curDir, true)
				if err != nil {
					log.Fatalln("Failed to save file", err)
				}
			}
		} else {
			log.Println("Nothing found")
		}
	} else if GetGenre != 0 {
		result, err := store.GetGenre(GetGenre)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if GetSerie != 0 {
		result, err := store.GetSerie(GetSerie)
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if Stat {
		result, err := store.GetSummary()
		if err == nil {
			utils.PrintJson(result, true)
		} else {
			log.Println("Nothing found")
		}
	} else if About {
		devinfo := models.DevInfo{}
		devinfo.Author = "Alexandr Mikhailenko a.k.a Alex M.A.K."
		devinfo.Email = "alex-m.a.k@yandex.kz"
		devinfo.Project.Name = "mOPDS"
		devinfo.Project.Version = "0.1.0"
		devinfo.Project.Link = "github.com/mak-alex/mopds"
		devinfo.Project.Created = "24.03.18 22:59"

		utils.PrintJson(devinfo, true)
	} else {
		rest.NewRestService(Listen, store, Catalog).StartListen()
	}
}
