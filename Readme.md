# mOPDS API

Free program for Linux operating systems, designed to quickly create an electronic OPDS-catalog books. OPDS (Open Publication Distribution System) catalog allows you to access Your library via the Internet from most devices for reading electronic books, tablets, smartphones, etc.
### Current features:

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
	
```Usage: mopds [--catalog CATALOG] [--page PAGE] [--perpage PERPAGE] [--get-author GET-AUTHOR] [--get-book GET-BOOK] [--get-genre GET-GENRE] [--get-serie GET-SERIE] [--get-authors] [--get-books] [--get-genres] [--get-series] [--get-books-by-author GET-BOOKS-BY-AUTHOR] [--get-books-by-genre GET-BOOKS-BY-GENRE] [--get-books-by-serie GET-BOOKS-BY-SERIE] [--search-author SEARCH-AUTHOR] [--search-book SEARCH-BOOK] [--search-genre SEARCH-GENRE] [--search-serie SEARCH-SERIE] [--listen LISTEN] [--parse] [--save] [--search-lib-id SEARCH-LIB-ID] [--search-title SEARCH-TITLE] [--host HOST] [--username USERNAME] [--password PASSWORD] [--database DATABASE] [--dbtype DBTYPE] [--dblog] [--sslmode SSLMODE] [--logfile LOGFILE] [--verbose]```

### Options:
```
  --catalog CATALOG, -c CATALOG
                         Directory of library (mandatory)
  --page PAGE            Page navigation
  --perpage PERPAGE, -l PERPAGE
                         Limit results (-1 for no limit) (default 10)
  --get-author GET-AUTHOR
                         Get author by id
  --get-book GET-BOOK    Get book by id
  --get-genre GET-GENRE
                         Get genre by id
  --get-serie GET-SERIE
                         Get serie by id
  --get-authors          List all authors
  --get-books            List all books
  --get-genres           List all genres
  --get-series           List all series
  --get-books-by-author GET-BOOKS-BY-AUTHOR
                         List all author's books by id
  --get-books-by-genre GET-BOOKS-BY-GENRE
                         List all genre's books by id
  --get-books-by-serie GET-BOOKS-BY-SERIE
                         List all serie's books by id
  --search-author SEARCH-AUTHOR
                         Search authors, or books by author if comes with search-title
  --search-book SEARCH-BOOK
                         Search book by title or filename
  --search-genre SEARCH-GENRE
                         Search genre by genre name or section, or subsection
  --search-serie SEARCH-SERIE
                         Search serie by serie name
  --listen LISTEN        Set server listen address:port (default  [default: :8000]
  --parse, -p            Parse inpx to the local database
  --save, -s             Save book file to the disk
  --search-lib-id SEARCH-LIB-ID
                         Search book(s) by its libId
  --search-title SEARCH-TITLE
                         Search books by their title
  --host HOST [default: localhost]
  --username USERNAME [default: mopds]
  --password PASSWORD [default: mopds]
  --database DATABASE [default: mopds]
  --dbtype DBTYPE, -t DBTYPE
                         type db: sqlite/postgres/mysql (mandatory) [default: sqlite3]
  --dblog [default: true]
  --sslmode SSLMODE       only for postges connection
  --logfile LOGFILE
  --verbose
  --help, -h             display this help and exit
```

### Structure Project
```
mOPDS
├── datastore
│   ├── datastore.go  
│   └── datastorer.go
├── inpx
│   ├── archive-utils.go
│   ├── genre.json
│   └── inpxreader.go
├── LICENSE
├── Makefile
├── models
│   ├── annnotation.go
│   ├── author.go
│   ├── book.go
│   ├── catalog.go
│   ├── cover.go
│   ├── dbconfig.go
│   ├── genre.go
│   ├── paginate.go
│   ├── search.go
│   └── serie.go
├── mopds.go
├── paginate
│   └── paginate.go
├── rest
│   ├── restserver.go
│   └── restservice.go
├── utils
│   └── utils.go
```

### Dependencies
* github.com/alexflint/go-arg
* github.com/emicklei/go-restful
* github.com/jinzhu/gorm
