package books

import (
  "io"
  "os"
  "fmt"
  "log"
  // "reflect"
  "bufio"
  "strconv"
  "runtime"
  "strings"
  "io/ioutil"
  "archive/zip"
  "path/filepath"
  "bitbucket.org/enlab/mopds/models"
  "bitbucket.org/enlab/mopds/modules/fb2parse"
  "bitbucket.org/enlab/mopds/modules/datastore"
  "bitbucket.org/enlab/mopds/utils"
)

var (
  inpx string
  versionFile  = "version.info"
  numProcesses = runtime.NumCPU()
  books_added uint64 = 0
  books_skipped uint64 = 0
  books_deleted uint64 = 0
  arch_scanned uint64 = 0
  arch_skipped uint64 = 0
  bad_archives uint64 = 0
  bad_books uint64 = 0
  books_in_archives uint64 = 0
  mOPDS_SCAN_INPX bool = true
)

func LogStats() {
  log.Println(fmt.Sprintf("Books added:       %d", books_added))
  log.Println(fmt.Sprintf("Books skipped:     %d", books_skipped))
  log.Println(fmt.Sprintf("Bad books:         %d", bad_books))
  // if config.SOPDS_DELETE_LOGICAL {
      // log.Println(fmt.Sprintf("Books deleted    : ", books_deleted))
  // } else {
    // log.Println(fmt.Sprintf("Books DB entries deleted : ", books_deleted))
  // }
  log.Println(fmt.Sprintf("Books in archives: %d", books_in_archives))
  log.Println(fmt.Sprintf("Archives scanned : %d", arch_scanned))
  log.Println(fmt.Sprintf("Archives skipped : %d", arch_skipped))
  log.Println(fmt.Sprintf("Bad archives     : %d", bad_archives))
}

func get_files(dataDir string, format string) ([]string) {
	files, err := filepath.Glob(filepath.Join(dataDir, format))
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func ProcessALL(dataDir string, store datastore.DataStorer) {
	genres, _ := utils.LoadGenres("conf/genre.json")
  if mOPDS_SCAN_INPX {
    inpxErr := processINPX(dataDir, genres, store)
    if inpxErr != nil {
      fmt.Println(inpxErr)
    }
  } else {
    zipsErr := processZIPS(dataDir, genres, store)
    if zipsErr != nil {
      fmt.Println(zipsErr)
    }
    go func() {
      processFILES(dataDir, genres, store)
    }()
  }
}

func processZIPS(dataDir string, genres utils.Genres, store datastore.DataStorer) (err error) {
  zip_files := get_files(dataDir, "*.zip")
  if len(zip_files) > 0 {
    for _, dataFile := range zip_files {
      r, err := zip.OpenReader(dataFile)
      defer r.Close()
      if err != nil {
        return err
      }

      _, catalog_name := filepath.Split(dataFile)
      catalog_size := utils.GetSizeCatalog(dataFile)

      files := make(chan *zip.File, 10)
      done := make(chan struct{})
      results := make(chan *models.Book, 100)

      go func() {
        for _, file := range r.File {
          files <- file
        }
        close(files)
      }()

      log.Println("Paralleling in ", numProcesses)
      for i := 0; i < numProcesses; i++ {
        go processZIP(files, dataDir, catalog_name, catalog_size, genres, results, done, store)
      }

      waitAndProcessResults(done, results, store)
    }
  } else {
    return nil
  }
  return err
}


func processZIP(files <-chan *zip.File, dataDir string, catalog_name string, catalog_size int64, genres utils.Genres, results chan<- *models.Book, done chan<- struct{}, store datastore.DataStorer) {
  defer func() {
    done <- struct{}{}
  }()

  for file := range files {
    if !strings.HasSuffix(file.FileInfo().Name(), ".fb2") {
      continue
    }
    rc, err := file.Open()
    defer rc.Close()
    if err != nil {
      fmt.Println(err)
    }
    _, fname := filepath.Split(catalog_name)
    //extension := filepath.Ext(fname)

    update := false
    if store.IsFileExist(fname) {
      fmt.Println(fmt.Sprintf("File %s already exists in the db. Data will be updated...", fname))
      update = true
    }

    log.Println("Processing file ", file.FileInfo().Name())
    content, err := ioutil.ReadAll(rc)
    if err == nil {
      b, _ := fb2parse.ParseBook(content, genres, true)
      book, bookErr := processFILE(b, genres, file.FileInfo().Name(), strconv.FormatInt(file.FileInfo().Size(), 10), update, dataDir)
      if bookErr == nil {
        book.Catalog = models.Catalog{
          Path:    dataDir,
          CatName: catalog_name,
          CatSize: catalog_size,
        }
        book.Update = update
        books_added++
        results <- book
      } else {
        fmt.Println(bookErr)
      }
    } else {
      fmt.Println(err)
    }
    if err != nil {
      if err != io.EOF {
        log.Println("failed to finish reading the file:", err)
      }
      break
    }
  }
}

func processINPX(dataDir string, genres utils.Genres, store datastore.DataStorer) (err error) {
  inpx_files := get_files(dataDir, "*.inpx")
  if len(inpx_files) != 0 {
    for _, dataFile := range inpx_files {
      inpx = dataFile
      r, err := zip.OpenReader(dataFile)
      defer r.Close()
      if err != nil {
        return err
      }

      files := make(chan *zip.File, 10)
      done := make(chan struct{})
      results := make(chan *models.Book, 100)

      go func() {
        for _, file := range r.File {
          files <- file
        }
        close(files)
      }()

      log.Println("Paralleling in ", numProcesses)
      for i := 0; i < numProcesses; i++ {
        go processINP(files, dataDir, genres, results, done, store)
      }

      waitAndProcessResults(done, results, store)
    }
    store.Close()
  }

	return nil
}

func processINP(files <-chan *zip.File, dataDir string, genres utils.Genres, results chan<- *models.Book, done chan<- struct{}, store datastore.DataStorer) {
	defer func() {
		done <- struct{}{}
	}()

	for file := range files {
		if !strings.HasSuffix(file.FileInfo().Name(), ".inp") {
			continue
		}

		zipCatalog := strings.Replace(file.FileInfo().Name(), ".inp", ".zip", -1)

		update := false
		if store.IsCatalogExist(zipCatalog) {
			log.Printf("Catalog %s already exists in the db. Data will be updated...", file.FileInfo().Name())
			update = true
		}

		log.Println("Processing file ", file.FileInfo().Name())
		rc, err := file.Open()
		if err == nil {
			reader := bufio.NewReader(rc)
			for {
				line, err := reader.ReadString('\n')
				if line != "" {
					book, bookErr := processBOOK(line, genres, update, zipCatalog)
					if bookErr == nil {
            book.Catalog = models.Catalog{
              Path:    dataDir,
              CatName: strings.Replace(file.FileInfo().Name(), ".inp", ".zip", -1),
              CatSize: utils.GetSizeCatalog(filepath.Join(dataDir, strings.Replace(file.FileInfo().Name(), ".inp", ".zip", -1))),
            }
            book.Update = update
            books_added++
            results <- book
					} else {
						log.Println(bookErr)
					}
				}
				if err != nil {
					if err != io.EOF {
						log.Println("failed to finish reading the file:", err)
					}
					break
				}
			}

			rc.Close()
		}

		if err != nil {
			log.Println("Error occured while reading file", err)
		}
		log.Println("Done with ", file.FileInfo().Name())
	}

}

func processFILES(dataDir string, genres utils.Genres, store datastore.DataStorer) {
  book_files := get_files(dataDir, "*.fb2")
  if len(book_files) > 0 {
    log.Println("Start indexing of the *.fb2")
    done := make(chan struct{})
    results := make(chan *models.Book, 100)
    go func() {
      for _, book_file := range book_files {
        log.Println("Processing file ", book_file)
        xmlFile, _ := os.Open(book_file)
        defer xmlFile.Close()
        content, _ := ioutil.ReadAll(xmlFile)
        book_size := strconv.Itoa(utils.GetSizeFile(book_file))
        update := false
        _, fname := filepath.Split(book_file)
        extension := filepath.Ext(fname)
        fbname := strings.TrimRight(fname, extension)
        if store.IsFileExist(fbname) {
          log.Printf("File %s already exists in the db. Data will be updated...", fbname)
          update = true
        }
        b, _ := fb2parse.ParseBook(content, genres, true)
        book, bookErr := processFILE(b, genres, book_file, book_size, update, dataDir)
        if bookErr == nil {
          bsize, _ := strconv.ParseInt(book_size, 10, 64)
          book.Catalog = models.Catalog{
            Path:    dataDir,
            CatName: fname,
            CatSize: bsize,
          }
          book.Update = update
          books_added++
          results <- book
        }
      }
      return
    }()
    waitAndProcessResults(done, results, store)
    store.Close()
  }
}

func processFILE(b fb2parse.BookInfo, genres utils.Genres, filename string, filesize string, update bool, dataDir string) (book *models.Book, err error) {
  book = new(models.Book)

  if !update {
    book.Authors = b.Authors
    for _, g := range b.Genres {
      book.Genres = append(book.Genres, models.Genre{Genre: g.Genre, Section: g.Section, Subsection: g.Subsection})
    }
    if len(b.Sequence) > 0 {
      book.Series = append(book.Series, models.Serie{Ser: b.Sequence, SearchSer: strings.ToUpper(b.Sequence)})
    }
    book.Annotations = append(book.Annotations, models.Annotation{Value: b.Annotation})
    for _, c := range b.Covers {
      book.Covers = append(book.Covers, c)
    }
  }
  path, fname := filepath.Split(filename)
  extension := filepath.Ext(fname)
  fbname := strings.TrimRight(fname, extension)
	book.Title = strings.ToLower(b.Title)
	book.SerNo = b.SerNo
  book.FileName = fbname
  book.FileSize = filesize
	// book.LibID = elements[7]
	// book.Del = elements[8]
  book.Format = extension
	book.DocDate = b.DocDate
	book.Lang = b.Language
  book.LangCode = utils.GetLangCode(strings.ToLower(b.Title))
  book.Avail = 2
  book.Path = path //filepath.Join(inpx, strings.Replace(zipCatalog, ".zip", ".inp", -1))

  return book, nil
}

func processBOOK(line string, genres utils.Genres, update bool, zipCatalog string) (book *models.Book, err error) {
	elements := strings.Split(line, string([]byte{0x04}))
	if len(elements) < 12 {
		return book, fmt.Errorf("Illegal number of elements")
	}

	book = new(models.Book)
	if !update {
		for _, author := range utils.TrimSlice(strings.Split(elements[0], ":")) {
			author = strings.TrimSpace(strings.Replace(author, ",", " ", -1))
			book.Authors = append(book.Authors, models.Author{FullName: strings.ToLower(author), SearchFullName: strings.ToUpper(author)})
		}
		for _, genre := range utils.TrimSlice(strings.Split(elements[1], ":")) {
			g := genres.GetDetail(genre)
			book.Genres = append(book.Genres, models.Genre{Genre: g.Genre, Section: g.Section, Subsection: g.Subsection})
		}
		if len(elements[3]) > 0 {
			book.Series = append(book.Series, models.Serie{Ser: elements[3], SearchSer: strings.ToUpper(elements[3])})
		}
	}
	book.Title = strings.ToLower(elements[2])
	book.SerNo = elements[4]
	book.FileName = elements[5]
	book.FileSize = elements[6]
	book.LibID = elements[7]
	book.Del = elements[8]
	book.Format = elements[9]
	book.DocDate = elements[10]
	book.Lang = elements[11]
  book.LangCode = utils.GetLangCode(strings.ToLower(elements[2]))
  book.Avail = 2
  book.Path = filepath.Join(inpx, strings.Replace(zipCatalog, ".zip", ".inp", -1))
  // utils.PrintJson(book, true)

	return book, nil
}

func waitAndProcessResults(done <-chan struct{}, results <-chan *models.Book, store datastore.DataStorer) {
	for working := numProcesses; working > 0; {
		select {
		case book := <-results:
			if book.Update {
        store.UpdateBook(book)
			} else {
				store.PutBook(book)
			}
		case <-done:
			log.Println("Gorutine finished work")
			working--
		}
	}

	for {
		select { // Nonblocking
		case book := <-results:
			if book.Update {
        store.UpdateBook(book)
			} else {
        store.PutBook(book)
			}
		default:
			return
		}
	}
}
