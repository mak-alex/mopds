package inpx

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"bitbucket.org/enlab/mopds/datastore"
	"bitbucket.org/enlab/mopds/models"
	"github.com/centrypoint/fb2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Genres []struct {
	Genre      string `json:"genre"`
	Section    string `json:"section"`
	Subsection string `json:"subsection"`
}

type Genre struct {
	Genre      string `json:"genre"`
	Section    string `json:"section"`
	Subsection string `json:"subsection"`
}

var (
	versionFile  = "version.info"
	numProcesses = runtime.NumCPU()
	genres       Genres
)

func getSizeCatalog(catalog string) int64 {
	fi, e := os.Stat(catalog);
	if e != nil {
	    fmt.Println(e)
	}
	// get the size
	return fi.Size()
}

func LoadGenres(file string) (Genres, error) {
	if genres == nil {
		fmt.Println("Opened genres list")
		genresFile, err := os.Open(file)
		defer genresFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		jsonParser := json.NewDecoder(genresFile)
		jsonParser.Decode(&genres)
	}

	return genres, nil
}

func (genres Genres) GetDetail(genre_code string) Genre {
	var genre Genre

	for i := 0; i < len(genres); i++ {
		if genre_code == genres[i].Genre {
			genre = genres[i]
		}
	}

	return genre
}

func getDescriptionAndCover(dataDir string, book *models.Book) (err error) {
	container := book.Catalog.CatName
	fileName := book.FileName + "." + book.Format

	r, err := zip.OpenReader(filepath.Join(dataDir, container))
	if err != nil {
		log.Printf("Failed to open container %s\n", container)
		return fmt.Errorf("Failed to open container %s", container)
	}
	defer r.Close()

	for _, file := range r.File {
		if file.FileInfo().Name() == fileName {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			content, err := ioutil.ReadAll(rc)
			p := fb2.New(content)
			result, err := p.Unmarshal()
			if err != nil {
				fmt.Println(err)
			}
			annotation := strings.Join(result.Description.TitleInfo.Annotation, " ")
			book.Annotations = append(book.Annotations, models.Annotation{Value: annotation})
			covers := result.Binary
			for i := 0; i < len(covers); i++ {
				if strings.Contains(covers[i].ID, "cover") {
					book.Covers = append(book.Covers, models.Cover{Name: covers[i].ID, Value: covers[i].Value, ContentType: covers[i].ContentType})
				}
			}

			break
		}
	}

	return err
}

func ReadInpxFile(dataFile string, dataDir string, store datastore.DataStorer) (err error) {
	r, err := zip.OpenReader(dataFile)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer r.Close()

	files := make(chan *zip.File, 10)
	done := make(chan struct{})
	results := make(chan *models.Book, 100)

	go func() {
		for _, file := range r.File {
			files <- file
		}
		close(files)
	}()

	log.Println("Paralleling in ", numProcesses*2)
	for i := 0; i < numProcesses; i++ {
		go processInp(files, dataDir, results, done, store)
	}

	waitAndProcessResults(done, results, store)

	store.Close()

	return nil
}

func processInp(files <-chan *zip.File, dataDir string, results chan<- *models.Book, done chan<- struct{}, store datastore.DataStorer) {

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
					book, bookErr := processBook(line, update, zipCatalog)
					if bookErr == nil {
						book.Catalog = models.Catalog{
							Path: dataDir,
							CatName: strings.Replace(file.FileInfo().Name(), ".inp", ".zip", -1),
							CatSize: getSizeCatalog(filepath.Join(dataDir, strings.Replace(file.FileInfo().Name(), ".inp", ".zip", -1)))}
						getDescriptionAndCover(dataDir, book)
						book.Update = update
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

func trimSlice(in []string) []string {
	for len(in) > 0 && in[len(in)-1] == "" {
		in = in[:len(in)-1]
	}
	return in
}

func processBook(line string, update bool, zipCatalog string) (book *models.Book, err error) {
	elements := strings.Split(line, string([]byte{0x04}))
	genres, _ := LoadGenres("/home/mak/.go/src/bitbucket.org/enlab/mopds/inpx/genre.json")
	if len(elements) < 12 {
		return book, fmt.Errorf("Illegal number of elements")
	}

	book = new(models.Book)
	if !update {
		for _, author := range trimSlice(strings.Split(elements[0], ":")) {
			author = strings.TrimSpace(strings.Replace(author, ",", " ", -1))
			book.Authors = append(book.Authors, models.Author{FullName: strings.ToLower(author), SearchFullName: strings.ToUpper(author)})
		}
		for _, genre := range trimSlice(strings.Split(elements[1], ":")) {
			g := genres.GetDetail(genre)
			book.Genres = append(book.Genres, models.Genre{Genre: g.Genre, Section: g.Section, Subsection: g.Subsection})
		}
		if len(elements[3]) > 0 {
			book.Series = append(book.Series, models.Serie{Ser: elements[3], SearchSer: strings.ToUpper(elements[3])})
		}
	}
	book.Title = strings.ToLower(elements[2])
	book.SearchTitle = strings.ToUpper(elements[2])
	book.SerNo = elements[4]
	book.FileName = elements[5]
	book.FileSize = elements[6]
	book.LibID = elements[7]
	book.Del = elements[8]
	book.Format = elements[9]
	book.DocDate = elements[10]
	book.Lang = elements[11]

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
