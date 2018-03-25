package rest

import (
	"archive/zip"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/modules/datastore"
	"bitbucket.org/enlab/mopds/modules/inpx"
	"bitbucket.org/enlab/mopds/utils"
	"github.com/emicklei/go-restful"
)

/*
TODO: доработать поиск
TODO: добавить конвертацию в epub
TODO: во-вторых, рест надо такой:
	/api/{authors,genres,series}/1?fields=[list]
	/api/books/1/download?format=[string] // string - fb2 or epub
*/

type RestService struct {
	listen    string
	dataDir   string
	dataStore datastore.DataStorer
	container *restful.Container
}

func (service RestService) registerBookResource(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/book").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/search").
		To(service.searchBooks).
		Doc("Search for the books").
		Operation("searchBooks").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.POST("/series").
		To(service.searchSeries).
		Doc("Search for the books").
		Operation("searchBooks").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/lib/{libId}").
		To(service.getBooksByLibID).
		Doc("Get books by libId").
		Operation("getBooksByLibId").
		Param(ws.PathParameter("libId", "libId of the book").DataType("string")).
		Returns(200, "OK", []models.Book{}))

	container.Add(ws)
}

func (service RestService) registerAuthorResource(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/about").
		To(service.getAbout).
		Doc("About author and project").
		Operation("getAbout").
		Returns(200, "OK", []models.DevInfo{}))

	ws.Route(ws.GET("/stat").
		To(service.getSummary).
		Doc("Get stat for catalog").
		Operation("getSummary").
		Returns(200, "OK", []models.Summary{}))

	ws.Route(ws.GET("/series").
		To(service.getSeries).
		Doc("Get all available series").
		Operation("getSeries").
		Returns(200, "OK", []models.Serie{}))

	ws.Route(ws.POST("/series/search").
		To(service.searchBooks).
		Doc("Search for the books").
		Operation("searchBooks").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/series/{serieId}").
		To(service.getSerie).
		Doc("Get serie info").
		Operation("getAuthor").
		Param(ws.PathParameter("serieId", "identifier of the serie").DataType("int")).
		Returns(200, "OK", models.Serie{}))

	ws.Route(ws.GET("/series/{serieId}/books").
		To(service.listSeriesBooks).
		Doc("Show author's books").
		Operation("listSeriesBooks").
		Param(ws.PathParameter("serieId", "identifier of the serie").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/series/{serieId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/series/{serieId}/books/langs").
		To(service.getLangs).
		Doc("Get all available books languages").
		Operation("getLangs").
		Returns(200, "OK", []string{"en"}))

	ws.Route(ws.GET("/series/{serieId}/books/{bookId}/download").
		To(service.downloadBook).
		Doc("Download book content").
		Operation("downloadBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/series/{serieId}/books/archive").
		To(service.downloadBooksArchive).
		Doc("Download books in single zip file").
		Operation("downloadBooksArchive").
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/genres").
		To(service.getGenres).
		Doc("Get all available genres").
		Operation("getGenres").
		Returns(200, "OK", []models.Genre{}))

	ws.Route(ws.GET("/genres/{genreId}").
		To(service.getGenre).
		Doc("Get genre info").
		Operation("getGenre").
		Param(ws.PathParameter("genreId", "identifier of the genre").DataType("int")).
		Returns(200, "OK", models.Genre{}))

	ws.Route(ws.GET("/genres/{genreId}/books").
		To(service.listGenresBooks).
		Doc("Show genre's books").
		Operation("listGenresBooks").
		Param(ws.PathParameter("genreId", "identifier of the genre").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/genres/{genreId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/genres/{genreId}/books/langs").
		To(service.getLangs).
		Doc("Get all available books languages").
		Operation("getLangs").
		Returns(200, "OK", []string{"en"}))

	ws.Route(ws.GET("/genres/{genreId}/books/{bookId}/download").
		To(service.downloadBook).
		Doc("Download book content").
		Operation("downloadBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/genres/{genreId}/books/archive").
		To(service.downloadBooksArchive).
		Doc("Download books in single zip file").
		Operation("downloadBooksArchive").
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/books").
		To(service.getBooks).
		Doc("Get all available books").
		Operation("getBooks").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/books/langs").
		To(service.getLangs).
		Doc("Get all available books languages").
		Operation("getLangs").
		Returns(200, "OK", []string{"en"}))

	ws.Route(ws.GET("/books/{bookId}/download").
		To(service.downloadBook).
		Doc("Download book content").
		Operation("downloadBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/books/archive").
		To(service.downloadBooksArchive).
		Doc("Download books in single zip file").
		Operation("downloadBooksArchive").
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/authors").
		To(service.searchAuthors).
		Doc("Search authors").
		Operation("searchAuthors").
		Returns(200, "OK", []models.Author{}))

	ws.Route(ws.GET("/authors/{authorId}").
		To(service.getAuthor).
		Doc("Get author's info").
		Operation("getAuthor").
		Param(ws.PathParameter("authorId", "identifier of the author").DataType("int")).
		Returns(200, "OK", models.Author{}))

	ws.Route(ws.GET("/authors/{authorId}/books").
		To(service.listAuthorsBooks).
		Doc("Show author's books").
		Operation("listAuthorsBooks").
		Param(ws.PathParameter("authorId", "identifier of the author").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/authors/{authorId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/authors/{authorId}/books/langs").
		To(service.getLangs).
		Doc("Get all available books languages").
		Operation("getLangs").
		Returns(200, "OK", []string{"en"}))

	ws.Route(ws.GET("/authors/{authorId}/books/{bookId}/download").
		To(service.downloadBook).
		Doc("Download book content").
		Operation("downloadBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.GET("/authors/{authorId}/books/archive").
		To(service.downloadBooksArchive).
		Doc("Download books in single zip file").
		Operation("downloadBooksArchive").
		Returns(200, "OK", models.Book{}))

	ws.Route(ws.POST("/authors/search").
		To(service.searchAuthors).
		Doc("Search authors").
		Operation("searchAuthors").
		Returns(200, "OK", []models.Author{}))

	container.Add(ws)
}

func (service RestService) getSeries(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetSeries(page, per_page)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getGenres(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetGenres(page, per_page)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getAbout(request *restful.Request, response *restful.Response) {
	devinfo := models.DevInfo{}
	devinfo.Author = "Alexandr Mikhailenko a.k.a Alex M.A.K."
	devinfo.Email = "alex-m.a.k@yandex.kz"
	devinfo.Project.Name = "mOPDS"
	devinfo.Project.Version = "0.1.0"
	devinfo.Project.Link = "bitbucket.org/enlab/mopds"
	devinfo.Project.Created = "24.03.18 22:59"
	response.WriteEntity(devinfo)
}

func (service RestService) getSummary(request *restful.Request, response *restful.Response) {
	result, err := service.dataStore.GetSummary()
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getBooks(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetBooks(search.Title, page, per_page)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getBook(request *restful.Request, response *restful.Response) {
	bookID, _ := strconv.ParseUint(request.PathParameter("bookId"), 0, 32)
	log.Println("Requesting book ", bookID)
	result, err := service.dataStore.GetBook(uint(bookID))
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getBooksByLibID(request *restful.Request, response *restful.Response) {
	libID := request.PathParameter("libId")
	log.Println("Get books by libId ", libID)
	result, err := service.dataStore.FindBooksByLibID(libID)
	if err == nil && len(result) != 0 {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Nothing was found\n")
	}
}

func (service RestService) downloadBook(request *restful.Request, response *restful.Response) {
	bookID, _ := strconv.ParseUint(request.PathParameter("bookId"), 0, 32)
	log.Println("Downloading book ", bookID)
	result, err := service.dataStore.GetBook(uint(bookID))
	if err == nil {
		outName := result.GetFullFilename()

		response.AddHeader("Content-Type", "application/octet-stream")
		response.AddHeader("Content-disposition", "attachment; filename*=UTF-8''"+strings.Replace(url.QueryEscape(
			utils.ReplaceUnsupported(outName)), "+", "%20", -1))

		err := inpx.UnzipBookToWriter(service.dataDir, result, response)
		if err != nil {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
		}
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) downloadBooksArchive(request *restful.Request, response *restful.Response) {
	request.Request.ParseForm()
	ids := request.Request.Form["id"]
	if len(ids) > 0 {
		response.Header().Set("Content-Type", "application/zip")
		response.Header().Set("Content-disposition", "attachment; filename*=UTF-8''"+strings.Replace(url.QueryEscape(
			"mopds-"+time.Now().Format("2006-01-02T15-04-05")+".zip"), "+", "%20", -1))
		zipWriter := zip.NewWriter(response)

		idsChan := make(chan string)
		done := make(chan bool)

		go func() {
			for {
				id, more := <-idsChan
				if more {
					bookID, _ := strconv.ParseUint(id, 0, 32)
					book, err := service.dataStore.GetBook(uint(bookID))
					if err == nil {
						zipHeader := &zip.FileHeader{Name: book.GetFullFilename(), Method: zip.Deflate, Flags: 0x800}
						entry, err := zipWriter.CreateHeader(zipHeader)
						// entry, err := zipWriter.Create(book.GetFullFilename())

						if err == nil {
							inpx.UnzipBookToWriter(service.dataDir, book, entry)
						} else {
							log.Println("Failed to compress ", book.GetFullFilename())
						}
					} else {
						log.Println("Failed to get book ", id)
					}
				} else {
					done <- true
					return
				}
			}
		}()
		for _, id := range ids {
			idsChan <- id
		}
		close(idsChan)
		<-done
		zipWriter.Close()
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "No parameters passed\n")
	}
}

func (service RestService) searchBooks(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Searching books ", search)

	result, err := service.dataStore.GetBooks(search.Title, page, per_page)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Nothing was found\n")
	}
}

func (service RestService) searchSeries(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	log.Println("Searching book series ", search)

	result, err := service.dataStore.FindBooksSeries(search)
	if err == nil && len(result) != 0 {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Nothing was found\n")
	}
}

func (service RestService) getLangs(request *restful.Request, response *restful.Response) {
	log.Println("Getting languages")

	result, err := service.dataStore.GetLangs()
	if err == nil && len(result) != 0 {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Nothing was found\n")
	}
}

func (service RestService) searchAuthors(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	result, err := service.dataStore.GetAuthors(search.Author, page, per_page)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Nothing was found\n")
	}
}

func (service RestService) getGenre(request *restful.Request, response *restful.Response) {
	genreId, _ := strconv.ParseUint(request.PathParameter("genreId"), 0, 32)
	log.Println("Requesting genre ", genreId)

	result, err := service.dataStore.GetGenre(uint(genreId))
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No genre was found\n")
	}
}

func (service RestService) getSerie(request *restful.Request, response *restful.Response) {
	serieId, _ := strconv.ParseUint(request.PathParameter("serieId"), 0, 32)
	log.Println("Requesting serie ", serieId)

	result, err := service.dataStore.GetSerie(uint(serieId))
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No serie was found\n")
	}
}

func (service RestService) getAuthor(request *restful.Request, response *restful.Response) {
	authorId, _ := strconv.ParseUint(request.PathParameter("authorId"), 0, 32)
	log.Println("Requesting author ", authorId)

	result, err := service.dataStore.GetAuthor(uint(authorId))
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No author was found\n")
	}
}

func (service RestService) listGenresBooks(request *restful.Request, response *restful.Response) {
	genreId, _ := strconv.ParseUint(request.PathParameter("genreId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting genre's books ", genreId)

	result, err := service.dataStore.ListGenreBooks(uint(genreId), noDetails, page, per_page, models.Search{})
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) listSeriesBooks(request *restful.Request, response *restful.Response) {
	serieId, _ := strconv.ParseUint(request.PathParameter("serieId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting serie's books ", serieId)

	result, err := service.dataStore.ListSerieBooks(uint(serieId), noDetails, page, per_page, models.Search{})
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) listAuthorsBooks(request *restful.Request, response *restful.Response) {
	authorId, _ := strconv.ParseUint(request.PathParameter("authorId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting author's books ", authorId)

	result, err := service.dataStore.ListAuthorBooks(uint(authorId), noDetails, page, per_page, models.Search{})
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) listAuthorsBooksPost(request *restful.Request, response *restful.Response) {
	authorId, _ := strconv.ParseUint(request.PathParameter("authorId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	search := models.Search{}
	request.ReadEntity(&search)

	log.Println("Requesting author's books ", authorId)

	result, err := service.dataStore.ListAuthorBooks(uint(authorId), noDetails, page, per_page, search)
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) StartListen() {
	log.Println("Start listening on ", service.listen)
	server := &http.Server{Addr: service.listen, Handler: service.container}
	log.Fatal(server.ListenAndServe())
}

func NewRestService(listen string, dataStore datastore.DataStorer, dataDir string) RestServer {
	service := new(RestService)
	service.listen = listen
	service.dataStore = dataStore
	service.dataDir = dataDir
	service.container = restful.NewContainer()
	service.container.Router(restful.CurlyRouter{})
	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-Total-Count"},
		AllowedHeaders: []string{"Content-Type", "Accept", "Content-Length", "X-Total-Count=100"},
		AllowedMethods: []string{"GET", "POST"},
		CookiesAllowed: false,
		Container:      service.container}
	service.container.Filter(cors.Filter)

	// Add container filter to respond to OPTIONS
	service.container.Filter(service.container.OPTIONSFilter)

	service.registerBookResource(service.container)
	service.registerAuthorResource(service.container)

	return service
}
