package main

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"path/filepath"
	"encoding/json"

	"bitbucket.org/enlab/mopds/datastore"
	"bitbucket.org/enlab/mopds/inpx"
	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/rest"
	"bitbucket.org/enlab/mopds/utils"
	"github.com/alexflint/go-arg"
)
// TODO: перенести уже проект на bitbucket.org/enlab/mopds
// ибо хуй его знает, всяко бывает а без отката...пздц крч

/*
TODO: добить консольный опрос...
TODO: РЕШИТЬ ВОПРОС С ДУБЛЯМИ (СЕРИЯ/КНИГИ/ЖАНРЫ/АВТОРЫ) при индексации
TODO: во-первых, скан архивов это хорошо, а если в директории только файлы?
TODO: во-вторых, смотри rest/restservice.go
TODO: в третьих, необходим проход по всем найденным *.inpx...не один
TODO: в четвертых, необходимы условия сканирования INPX:
	1) если решили проверять на наличии ZIP файла или книги в ZIP,
	а самого файла нет, пропускаем обработку архива
	2) если решили проверять наличие файлов (fb2) в ZIP, то
	получаем однократно список файлов
		2.1) если книги не оказалось в наличии, пропускаем обработку
	3) если книга помечена как удаленная, пропускаем обработку
	4) если в имени автора нет запятой, то фамилию переносим в начало
TODO: в пятых, условия:
	1) если разрешена обработка inpx, то при нахождении обрабатываем его
	и прекращаем обработку текущего каталога
		1.1) пропускаем обработку файлов (zip) в текущем каталоге, если
		найдены inpx
TODO: в шестых, добавить колбек счетчик, чего сколько было добавлено,
	почему не добавленно и чего и т.д. и т.п., по анологии с sopds
TODO: в 7-ых, логировать в файл, также стоит подумать о глубине логирования и связи DBLog and (Verbose or Debug)
*/

type DBConfig struct {
	Host     string
	Username string
	Password string
	Database string
	DBType	 string `arg:"-t" help:"type db: sqlite/postgres/mysql (mandatory)"`
	DBLog    bool
	SSLMode	 string `help:" only for postges connection"`
}

type Mopds struct {
	Catalog	string  `arg:"-c" help:"Directory of library (mandatory)"`
	Page		int	`help:"Page navigation"`
	PerPage		int	`arg:"-l" help:"Limit results (-1 for no limit) (default 10)"`

	GetAuthor	uint	`arg:"--get-author" help:"Get author by id"`
	GetBook		uint	`arg:"--get-book" help:"Get book by id"`
	GetGenre	uint	`arg:"--get-genre" help:"Get genre by id"`
	GetSerie	uint	`arg:"--get-serie" help:"Get serie by id"`

	GetAuthors	bool	`arg:"--get-authors" help:"List all authors"`
	GetBooks	bool	`arg:"--get-books" help:"List all books"`
	GetGenres	bool	`arg:"--get-genres" help:"List all genres"`
	GetSeries	bool	`arg:"--get-series" help:"List all series"`

	GetBooksByAuthor	uint	`arg:"--get-books-by-author" help:"List all author's books by id"`
	GetBooksByGenre	uint	`arg:"--get-books-by-genre" help:"List all genre's books by id"`
	GetBooksBySerie	uint	`arg:"--get-books-by-serie" help:"List all serie's books by id"`

	SearchAuthor	string	`arg:"--search-author" help:"Search authors, or books by author if comes with search-title"`
	SearchBook	string	`arg:"--search-book" help:"Search book by title or filename"`
	SearchGenre	string	`arg:"--search-genre" help:"Search genre by genre name or section, or subsection"`
	SearchSerie	string	`arg:"--search-serie" help:"Search serie by serie name"`

	Listen		string	`arg:"--listen" help:"Set server listen address:port (default ":8000")"`
	Parse		bool	`arg:"-p" help:"Parse inpx to the local database"`
	Save		bool	`arg:"-s" help:"Save book file to the disk"`
	SearchLibID	string	`arg:"--search-lib-id" help:"Search book(s) by its libId"`
	SearchTitle	string	`arg:"--search-title" help:"Search books by their title"`
}

func (Mopds) Description() string {
	return `
mOPDS API

Free program for Linux operating systems, designed to quickly create an electronic OPDS-catalog books. OPDS (Open Publication Distribution System) catalog allows you to access Your library via the Internet from most devices for reading electronic books, tablets, smartphones, etc.
Current features:

    Recursive crawl specified in the directory configuration file. High speed of scanning-cataloging.
    Placement in the catalog of e-books, any formats specified in the configuration file.
    Extraction of meta information from ebooks FB2, EPUB, MOBI (title, authors, genres, series, abstract, language, editing date).
    The retrieval and display of covers of e-books FB2, EPUB, MOBI.
    Duplicate detection of books.
    Search books in zip-archives.
    Fast download collections of INPX file.
    There are five possible sorts in the catalog: by catalogs, by authors, by name, by genre, by series.
    Arbitrary search books by name, author name and series name.
    You can set a limit to whether items are displayed on a page.
    Archiving of books for downloading.
    Conversion of FB2 books to EPUB and MOBI "on the fly" using external converters.
    Database support SQLite, MySQL, PostgreSQL
	`
}

type LogOptions struct {
	LogFile string
	Verbose bool
}

func findINPX(catalog string) []string {
	inpx_files, err := filepath.Glob(filepath.Join(catalog, "*.inpx"))
	if err != nil {
		log.Fatal(err)
	}
	return inpx_files
}

func setConfig(dataDir string, settings DBConfig) *models.DBConfig {
	result := new(models.DBConfig)

	result.DBType = settings.DBType
	result.DBLog = settings.DBLog
	if settings.DBType == "sqlite3" {
		fileData, err := ioutil.ReadFile(filepath.Join(dataDir, "mopds_db.sqlite"))
		if err == nil {
			err = json.Unmarshal(fileData, result)
		}

		if err != nil { // fallback to sqlite
			result.DBParams = filepath.Join(dataDir, "mopds_db.sqlite")
		}
	} else if settings.DBType == "postgres" {
		if settings.SSLMode == "" {
			settings.SSLMode = "disable"
		}
		result.DBParams = fmt.Sprintf("user=%s password=%s DB.name=%s sslmode=%s", settings.Username, settings.Password, settings.Database, settings.SSLMode)
	}

	return result
}

func main() {
	curDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	var args struct {
		Mopds
		DBConfig
		LogOptions
	}
	if args.Mopds.Listen == "" {
		args.Mopds.Listen = ":8000"
	}
	if args.DBConfig.Host == "" {
		args.DBConfig.Host = "localhost"
	}
	if args.DBConfig.Username == "" {
		args.DBConfig.Username = "mopds"
	}
	if args.DBConfig.Password == "" {
		args.DBConfig.Password = "mopds"
	}
	if args.DBConfig.Database == "" {
		args.DBConfig.Database = "mopds"
	}
	if args.DBConfig.DBType == "" {
		args.DBConfig.DBType = "sqlite3"
	}
	args.DBConfig.DBLog = true


	arg.MustParse(&args)

	store, err := datastore.NewDBStore(setConfig(args.Mopds.Catalog, args.DBConfig))
	if err != nil {
		log.Fatalln("Failed to open database")
	}
	defer store.Close()

	if args.Mopds.Parse {
		inpx_file := findINPX(args.Mopds.Catalog)[0]
		log.Printf("Opening %s to parse data\n", args.Mopds.Parse)
		go func() {
			inpx.ReadInpxFile(inpx_file, args.Mopds.Catalog, store)
		}()
		rest.NewRestService(args.Mopds.Listen, store, args.Mopds.Catalog).StartListen()
	} else if args.Mopds.SearchLibID != "" {
		result, err := store.FindBooksByLibID(args.Mopds.SearchLibID)
		if err == nil && len(result) != 0 {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.SearchTitle != "" {
		result, err := store.FindBooks(models.Search{Title: args.Mopds.SearchTitle, Author: args.Mopds.SearchAuthor, Limit: args.Mopds.PerPage})
		if err == nil && len(result) != 0 {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.SearchAuthor != "" {
		result, err := store.GetAuthors(args.Mopds.SearchAuthor, args.Mopds.Page, args.Mopds.PerPage)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetAuthors {
		result, err := store.GetAuthors("", args.Mopds.Page, args.Mopds.PerPage)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetBooks {
		result, err := store.GetBooks("", args.Mopds.Page, args.Mopds.PerPage)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetGenres {
		result, err := store.GetGenres(args.Mopds.Page, args.Mopds.PerPage) // фильтр добавить
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetSeries {
		result, err := store.GetSeries(args.Mopds.Page, args.Mopds.PerPage) // фильтр добавить
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetBooksByAuthor != 0 {
		result, err := store.ListAuthorBooks(args.Mopds.GetBooksByAuthor, false, args.Mopds.Page, args.Mopds.PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetBooksByGenre != 0 {
		result, err := store.ListGenreBooks(args.Mopds.GetBooksByGenre, false, args.Mopds.Page, args.Mopds.PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetBooksBySerie != 0 {
		result, err := store.ListSerieBooks(args.Mopds.GetBooksBySerie, false, args.Mopds.Page, args.Mopds.PerPage, models.Search{})
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetAuthor != 0 {
		result, err := store.GetAuthor(args.Mopds.GetAuthor)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetBook != 0 {
		result, err := store.GetBook(args.Mopds.GetBook)
		if err == nil {
			utils.PrintJson(result)
			if args.Mopds.Save {
				err = inpx.UnzipBookFile(args.Mopds.Catalog, result, curDir, true)
				if err != nil {
					log.Fatalln("Failed to save file", err)
				}
			}
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetGenre != 0 {
		result, err := store.GetGenre(args.Mopds.GetGenre)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else if args.Mopds.GetSerie != 0 {
		result, err := store.GetSerie(args.Mopds.GetSerie)
		if err == nil {
			utils.PrintJson(result)
		} else {
			log.Println("Nothing found")
		}
	} else {
		rest.NewRestService(args.Mopds.Listen, store, args.Mopds.Catalog).StartListen()
	}
}
