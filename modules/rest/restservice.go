// set ts=2 shiftwidth=2 expandtab
package rest

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"bitbucket.org/enlab/mopds/models"
	"bitbucket.org/enlab/mopds/modules/books"
	"bitbucket.org/enlab/mopds/modules/datastore"
	"bitbucket.org/enlab/mopds/utils"
	"github.com/emicklei/go-restful"
	// "github.com/gorilla/mux"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/mattrobenolt/go-memcached"
)

var rootdir = "./ui"

type Cache map[string]*memcached.Item

func (c Cache) Get(key string) memcached.MemcachedResponse {
	if item, ok := c[key]; ok {
		if item.IsExpired() {
			delete(c, key)
		} else {
			return &memcached.ItemResponse{item}
		}
	}
	return nil
}

func (c Cache) Set(item *memcached.Item) memcached.MemcachedResponse {
	c[item.Key] = item
	return nil
}

func (c Cache) Delete(key string) memcached.MemcachedResponse {
	delete(c, key)
	return nil
}

/*
TODO: доработать поиск
TODO: добавить конвертацию в epub
TODO: во-вторых, рест надо такой:
	/api/{authors,genres,series}/1?fields=[list]
	/api/books/1/download?format=[string] // string - fb2 or epub
*/
type IncludeLibs struct {
	WebixCSS string
	WebixJS  string
	LogicJS  string
}

func jwtAuthentication(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	if err := jwtMiddleware.CheckJWT(resp.ResponseWriter, req.Request); err != nil {
		fmt.Printf("Authentication error: %v", err)
	}
	chain.ProcessFilter(req, resp)
}

func basicAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// usr/pwd = admin/admin
	u, p, ok := req.Request.BasicAuth()
	if !ok || u != "admin" || p != "admin" {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}
	chain.ProcessFilter(req, resp)
}

func simpleUI(req *restful.Request, resp *restful.Response) {
	p := &IncludeLibs{
		WebixCSS: "/static/assets/js/webix/skins/contrast.css",
		WebixJS:  "/static/assets/js/webix/webix.js",
		LogicJS:  "/static/assets/js/mopds/main.js",
	}
	// you might want to cache compiled templates
	t, err := template.ParseFiles("ui/index.html")
	if err != nil {
		log.Fatalf("Template gave: %s", err)
	}
	t.Execute(resp.ResponseWriter, p)
}

type RestService struct {
	listen    string
	dataDir   string
	dataStore datastore.DataStorer
	container *restful.Container
}

func (service RestService) registerSimpleUIResource(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/").Filter(basicAuthenticate).To(simpleUI))
	ws.Route(ws.GET("/static/{subpath:*}").To(staticFromPathParam))
	ws.Route(ws.GET("/static").To(staticFromQueryParam))
	container.Add(ws)
}

func staticFromPathParam(req *restful.Request, resp *restful.Response) {
	actual := path.Join(rootdir, req.PathParameter("subpath"))
	fmt.Printf("serving %s ... (from %s)\n", actual, req.PathParameter("subpath"))
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		actual)
	// respe.WriteEntity(req.PathParameter)
}

func staticFromQueryParam(req *restful.Request, resp *restful.Response) {
	http.ServeFile(
		resp.ResponseWriter,
		req.Request,
		path.Join(rootdir, req.QueryParameter("resource")))
}

func (service RestService) registerApiResource(container *restful.Container) {
	restful.DefaultContainer.Router(restful.CurlyRouter{})
	ws := new(restful.WebService)
	ws.
		Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/langs").
		To(service.getLangs).
		Doc("Get all available books languages").
		Operation("getLangs").
		Returns(200, "OK", []string{"en"}))

	ws.Route(ws.GET("/authors").
		To(service.getAuthors).
		Doc("Search authors").
		Operation("getAuthors").
		Returns(200, "OK", []models.Author{}))

	ws.Route(ws.GET("/authors/{authorId}").
		To(service.getAuthor).
		Doc("Get author's info").
		Operation("getAuthor").
		Param(ws.PathParameter("authorId", "identifier of the author").DataType("int")).
		Returns(200, "OK", models.Author{}))

	ws.Route(ws.GET("/authors/{authorId}/books").
		To(service.getBooksByAuthorID).
		Doc("Show author's books").
		Operation("getBooksByAuthorID").
		Param(ws.PathParameter("authorId", "identifier of the author").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/authors/{authorId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

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
		To(service.getAuthors).
		Doc("Search authors").
		Operation("getAuthors").
		Returns(200, "OK", []models.Author{}))

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

	ws.Route(ws.POST("/books/search").
		To(service.getBooks).
		Doc("Search for the books").
		Operation("getBooks").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.POST("/books/series").
		To(service.getBooksBySerie).
		Doc("Search for the books from serie").
		Operation("getBooksBySerie").
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/books/lib/{libId}").
		To(service.getBooksByLibID).
		Doc("Get books by libId").
		Operation("getBooksByLibId").
		Param(ws.PathParameter("libId", "libId of the book").DataType("string")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/genres").
		To(service.getGenres).
		Doc("Get all available genres").
		Operation("getGenres").
		Returns(200, "OK", []models.Genre{}))

	ws.Route(ws.GET("/genres/menu").
		To(service.getGenresMenu).
		Doc("Format genres for WebixGroupList").
		Operation("getGenresMenu").
		Returns(200, "OK", []models.ItemMenu{}))

	ws.Route(ws.GET("/genres/{genreId}").
		To(service.getGenre).
		Doc("Get genre info").
		Operation("getGenre").
		Param(ws.PathParameter("genreId", "identifier of the genre").DataType("int")).
		Returns(200, "OK", models.Genre{}))

	ws.Route(ws.GET("/genres/{genreId}/books").
		To(service.getBooksByGenreID).
		Doc("Show genre's books").
		Operation("getBooksByGenreID").
		Param(ws.PathParameter("genreId", "identifier of the genre").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/genres/{genreId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

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

	ws.Route(ws.POST("/genres/search").
		To(service.getGenres).
		Doc("Search for the genres").
		Operation("getGenres").
		Returns(200, "OK", []models.Genre{}))

	ws.Route(ws.GET("/series").
		To(service.getSeries).
		Doc("Get all available series").
		Operation("getSeries").
		Returns(200, "OK", []models.Serie{}))

	ws.Route(ws.POST("/series/search").
		To(service.getSeries).
		Doc("Search for the series").
		Operation("getSeries").
		Returns(200, "OK", []models.Serie{}))

	ws.Route(ws.GET("/series/{serieId}").
		To(service.getSerie).
		Doc("Get serie info").
		Operation("getAuthor").
		Param(ws.PathParameter("serieId", "identifier of the serie").DataType("int")).
		Returns(200, "OK", models.Serie{}))

	ws.Route(ws.GET("/series/{serieId}/books").
		To(service.getBooksBySerieID).
		Doc("Show author's books").
		Operation("getBooksBySerieID").
		Param(ws.PathParameter("serieId", "identifier of the serie").DataType("int")).
		Returns(200, "OK", []models.Book{}))

	ws.Route(ws.GET("/series/{serieId}/books/{bookId}").
		To(service.getBook).
		Doc("Get specific book info").
		Operation("getBook").
		Param(ws.PathParameter("bookId", "identifier of the book").DataType("int")).
		Returns(200, "OK", models.Book{}))

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

	container.Add(ws)
}

func (service RestService) getSeries(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	log.Println("Searching serie ", search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetSeries(search.Series, page, per_page)

	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "Book wasn't found\n")
	}
}

func (service RestService) getGenresMenu(request *restful.Request, response *restful.Response) {
	result, err := service.dataStore.GetGenresMenu()
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
	log.Println("Searching genre ", search)
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetGenres(search.Genre, page, per_page)
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
	log.Println("Searching books ", search)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	random, _ := utils.ParseBool(request.QueryParameter("random"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	result, err := service.dataStore.GetBooks(search.Title, noDetails, random, page, per_page)
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
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))
	log.Println("Get books by libId ", libID)
	result, err := service.dataStore.GetBooksByLibID(libID, noDetails, page, per_page)
	if err == nil {
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

		err := books.UnzipBookToWriter(service.dataDir, result, response)
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
							books.UnzipBookToWriter(service.dataDir, book, entry)
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

func (service RestService) getBooksBySerie(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	log.Println("Searching book series ", search)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	result, err := service.dataStore.GetBooksBySerie(search.Title, search.Series, noDetails, page, per_page)
	if err == nil {
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

func (service RestService) getAuthors(request *restful.Request, response *restful.Response) {
	search := models.Search{}
	request.ReadEntity(&search)
	log.Println("Searching authors ", search)
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

func (service RestService) getBooksByGenreID(request *restful.Request, response *restful.Response) {
	genreId, _ := strconv.ParseUint(request.PathParameter("genreId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	random, _ := utils.ParseBool(request.QueryParameter("random"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting genre's books ", genreId)

	result, err := service.dataStore.GetBooksByGenreID(uint(genreId), noDetails, random, page, per_page, models.Search{})
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) getBooksBySerieID(request *restful.Request, response *restful.Response) {
	serieId, _ := strconv.ParseUint(request.PathParameter("serieId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	random, _ := utils.ParseBool(request.QueryParameter("random"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting serie's books ", serieId)

	result, err := service.dataStore.GetBooksBySerieID(uint(serieId), noDetails, random, page, per_page, models.Search{})
	if err == nil {
		response.WriteEntity(result)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "No books was found\n")
	}
}

func (service RestService) getBooksByAuthorID(request *restful.Request, response *restful.Response) {
	authorId, _ := strconv.ParseUint(request.PathParameter("authorId"), 0, 32)
	noDetails, _ := utils.ParseBool(request.QueryParameter("no-details"))
	random, _ := utils.ParseBool(request.QueryParameter("random"))
	page := utils.ParseInt(request.QueryParameter("page"))
	per_page := utils.ParseInt(request.QueryParameter("per_page"))

	log.Println("Requesting author's books ", authorId)

	result, err := service.dataStore.GetBooksByAuthorID(uint(authorId), noDetails, random, page, per_page, models.Search{})
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

	service.registerSimpleUIResource(service.container)
	service.registerApiResource(service.container)

	go func() {
		server := memcached.NewServer(":11211", make(Cache))
		log.Fatal(server.ListenAndServe())
	}()

	return service
}
