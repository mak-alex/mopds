# mOPDS API

Free program for Linux operating systems, designed to quickly create an electronic OPDS-catalog books. OPDS (Open Publication Distribution System) catalog allows you to access Your library via the Internet from most devices for reading electronic books, tablets, smartphones, etc.
## Current features:

* Recursive crawl specified in the directory configuration file. High speed of scanning-cataloging.
* Placement in the catalog of e-books, any formats specified in the configuration file.
* Extraction of meta information from ebooks FB2, EPUB, MOBI (title, authors, genres, series, abstract, language, editing date).
* The retrieval and display of covers of e-books FB2, EPUB, MOBI.
* Duplicate detection of books.
* Search books in zip-archives.
* Fast download collections of INPX file.
* There are five possible sorts in the catalog: by catalogs, by authors, by name, by genre, by series.
* Arbitrary search books by name, author name and series name.
* You can set a limit to whether items are displayed on a page.
* Archiving of books for downloading.
* Conversion of FB2 books to EPUB and MOBI "on the fly" using external converters.
* Database support SQLite, MySQL, PostgreSQL
	

## Install
```bash
mak@denied ~ $ go get bitbucket.org/enlab/mopds         # get project
mak@denied ~ $ cd $GOPATH/src/bitbucket.org/enlab/mopds # go to the directory
mak@denied ~ $ go get                                   # get depends
mak@denied ~ $ make run                                 # start project
mak@denied ~ $ make build                               # build project
mak@denied ~ $ make clean                               # clean project
```
## Usage of mopds:
```
  -about
    	About author and this project
  -catalog string
    	Directory of library (mandatory)
  -config string
    	Default configuration file (default "./conf/mopds.conf")
  -database string
    	Database name connect to database
  -dbtype string
    	Type used database: sqlite3, mysql or postgres (default "sqlite3")
  -get_author uint
    	Get author by id
  -get_authors
    	List all authors
  -get_book uint
    	Get book by id
  -get_books
    	List all books
  -get_books_by_author uint
    	List all author's books by id
  -get_books_by_genre uint
    	List all genre's books by id
  -get_books_by_serie uint
    	List all serie's books by id
  -get_genre uint
    	Get genre by id
  -get_genres
    	List all genres
  -get_serie uint
    	Get serie by id
  -get_series
    	List all series
  -host string
    	IP address for connect to database (default "modps")
  -listen string
    	Set server listen address:port (default ":8000")
  -page int
    	Pagination 1...n
  -parse
    	Parse inpx to the local database
  -password string
    	Password for connect to database (default "mopds")
  -per_page int
    	Limit results (-1 for no limit) (default 25)
  -save
    	Save book file to the disk (ex.: --get_book 1 --save)
  -search_author string
    	Search authors, or books by author if comes with search-title
  -search_book string
    	Search book by title or filename
  -search_genre string
    	Search genre by genre name or section, or subsection
  -search_lib_id string
    	Search book(s) by its libId
  -search_serie string
    	Search serie by serie name
  -search_title string
    	Search books by their title
  -sslmode string
    	Whether to use ssl mode or not, here's the question: disable or enable
  -stat
    	Book library statistics
  -username string
    	Username for connect to database (default "mopds")
  -verbose
    	Verbose output
```

## Example configuration file
```bash
page 1
per_page 25
listen :8000
host localhost
username mopds
password mopds
database mopds
dbtype postgres
sslmode disable
```

## Structure Project
```
mOPDS
├── conf
│   └── mopds.conf - configuration file
├── LICENSE
├── Makefile - run and build project	
├── models - database model
│   ├── annnotation.go
│   ├── author.go
│   ├── book.go
│   ├── catalog.go
│   ├── cover.go
│   ├── dbconfig.go
│   ├── devinfo.go
│   ├── genre.go
│   ├── paginate.go
│   ├── search.go
│   ├── serie.go
│   └── summary.go
├── modules - modules for working with database, pagination, inpx files, etc..
│   ├── datastore
│   │   ├── datastore.go
│   │   └── datastorer.go
│   ├── inpx
│   │   ├── archive-utils.go
│   │   ├── genre.json
│   │   └── inpxreader.go
│   ├── paginate
│   │   └── paginate.go
│   └── rest
│       ├── restserver.go
│       └── restservice.go
├── mopds.go 
├── Readme.md
└── utils
    └── utils.go - auxiliary functions, format conversion, etc.
```

## Example query
#### Get all authors
```bash
mak@denied ~ $ http :8000/authors
```
#### Get author by id
```bash
mak@denied ~ $ http :8000/authors/1
```
#### Get all books by author id
```bash
mak@denied ~ $ http :8000/authors/1/books
```
#### Get all books by author id and download bookd by id
```bash
mak@denied ~ $ http :8000/authors/1/books/1/download
```
#### Get all books
```bash
mak@denied ~ $ http :8000/books
```
#### Get book by id
```bash
mak@denied ~ $ http :8000/books/1
```
#### Download book by id
```bash
mak@denied ~ $ http :8000/books/1/download
```
#### Get all genres
```bash
mak@denied ~ $ http :8000/genres
```
#### Get genre by id
```bash
mak@denied ~ $ http :8000/genres/1
```
#### Get all books by genre id
```bash
mak@denied ~ $ http :8000/genres/1/books
```
#### Get all books by genre id and download bookd by id
```bash
mak@denied ~ $ http :8000/genres/1/books/1/download
```
#### Get all series
```bash
mak@denied ~ $ http :8000/series
```
#### Get serie by id
```bash
mak@denied ~ $ http :8000/series/1
```
#### Get all books by serie id
```bash
mak@denied ~ $ http :8000/series/1/books
```
#### Get all books by serie id and download bookd by id
```bash
mak@denied ~ $ http :8000/series/1/books/1/download
```

#### Start scanning the book library
```bash
mak@denied ~ $ http :8000/scan
```

#### Get configuration file
```bash
mak@denied ~ $ http :8000/conf

```
#### Update confiuration file
```bash
mak@denied ~ $ http --form POST :8000/conf data="catalog /mnt/books"
```
#### Find author by name
```bash
mak@denied ~ $ http POST :8000/api/v1/authors/search author=перумов
```
#### Find book by name
```bash
mak@denied ~ $ http POST :8000/api/v1/books/search title="Ген Химеры Часть 1"
```
#### Find genre by name
```bash
mak@denied ~ $ http POST :8000/api/v1/genres/search genre="Народные песни"
```
#### Find serie by name
```bash
mak@denied ~ $ http POST :8000/api/v1/series/search title="Ген Химеры"
```
#### Get random books by author id
```bash
mak@denied ~ $ http GET :8000/api/v1/authors/1/books?random=true&no-details=true
```
#### Get random books
```bash
mak@denied ~ $ http GET :8000/api/v1/books?random=true&no-details=true
```
#### Get random books by genre id
```bash
mak@denied ~ $ http GET :8000/api/v1/genres/1/books?random=true&no-details=true
```
#### Get random books by serie id
```bash
mak@denied ~ $ http GET :8000/api/v1/series/1/books?random=true&no-details=true
```

### Dependencies
* github.com/namsreal/flag
* github.com/emicklei/go-restful
* github.com/jinzhu/gorm

## TODO:

* add support for other book formats, epub/mobi/pdf/etc
* add the indexing process in the API output, etc.
* add opds catalog functionality
* Placement in the catalog of e-books, any formats specified in the configuration file.
