// set ts=2 shiftwidth=2 expandtab
package rest

import (
	// "archive/zip"
	// "fmt"
	"log"
	"net/http"
	"net/url"
	// "path"
	"strconv"
	"strings"
	// "text/template"
	"time"

	"github.com/StephanDollberg/go-json-rest-middleware-jwt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mak-alex/mopds/models"
	"github.com/mak-alex/mopds/modules/books"
	"github.com/mak-alex/mopds/modules/datastore"
	"github.com/mak-alex/mopds/utils"
	// "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var rootdir = "./ui"

const (
	DEFAULT_PAGE_SIZE int = 20
	MAX_PAGE_SIZE     int = 1000
)

type RestService struct {
	listen    string
	dataDir   string
	dataStore datastore.DataStorer
	container *rest.Api
}

func handle_auth(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(map[string]string{"authed": r.Env["REMOTE_USER"].(string)})
}

func (service RestService) StartListen() {
	log.Println("Start listening on ", service.listen)
	http.Handle("/api/v1/", http.StripPrefix("/api/v1", service.container.MakeHandler()))
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(rootdir))))
	log.Fatal(http.ListenAndServe(service.listen, nil))
}

func NewRestService(listen string, dataStore datastore.DataStorer, dataDir string) RestServer {
	jwt_middleware := &jwt.JWTMiddleware{
		Key:        []byte("secret key"),
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			return userId == "admin" && password == "admin"
		},
	}

	service := new(RestService)
	service.listen = listen
	service.dataStore = dataStore
	service.dataDir = dataDir
	service.container = rest.NewApi()
	service.container.Use(rest.DefaultDevStack...)
	service.container.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: jwt_middleware,
	})
	service.container.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			return origin == "http://localhost"
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "X-Custom-Header", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	statusMw := &rest.StatusMiddleware{}
	service.container.Use(statusMw)
	service.container.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/login", jwt_middleware.LoginHandler),
		rest.Get("/auth_test", handle_auth),
		rest.Get("/refresh_token", jwt_middleware.RefreshHandler),
		rest.Get("/.status",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(statusMw.GetStatus())
			},
		),
		rest.Get("/authors", service.getAuthors),
		rest.Get("/authors/:id", service.getAuthor),
		rest.Get("/authors/:id/books", service.getBooksByAuthorID),
		rest.Get("/books", service.getBooks),
		rest.Get("/books/:id", service.getBook),
		rest.Get("/books/:id/download", service.downloadBook),
		rest.Get("/genres", service.getGenres),
		rest.Get("/genres/:id", service.getGenre),
		rest.Get("/genres/:id/books", service.getBooksByGenreID),
		rest.Get("/series", service.getSeries),
		rest.Get("/series/:id", service.getSerie),
		rest.Get("/series/:id/books", service.getBooksBySerieID),
		rest.Get("/langs", service.getLangs),
	)
	if err != nil {
		log.Fatal(err)
	}
	service.container.SetApp(router)

	return service
}

func (service RestService) getAuthors(w rest.ResponseWriter, r *rest.Request) {
	search := models.Search{}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := service.dataStore.GetAuthors(search.Author, page, per_page)
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "authors not found", 400)
	}
}

func (service RestService) getAuthor(w rest.ResponseWriter, r *rest.Request) {
	authorID, _ := strconv.Atoi(r.PathParam("id"))

	result, err := service.dataStore.GetAuthor(uint(authorID))
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "author not found", 400)
	}
}

func (service RestService) getBooks(w rest.ResponseWriter, r *rest.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	no_details, _ := utils.ParseBool(r.URL.Query().Get("no-details"))
	random, _ := utils.ParseBool(r.URL.Query().Get("random"))

	if per_page <= 0 {
		per_page = DEFAULT_PAGE_SIZE
	}
	if per_page > MAX_PAGE_SIZE {
		per_page = MAX_PAGE_SIZE
	}
	if page == 0 {
		page = 1
	}

	result, err := service.dataStore.GetBooks("", no_details, random, page, per_page)
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "books not found", 400)
	}
}

func (service RestService) getBook(w rest.ResponseWriter, r *rest.Request) {
	bookID, _ := strconv.Atoi(r.PathParam("id"))
	result, err := service.dataStore.GetBook(uint(bookID))
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "book not found", 400)
	}
}

func (service RestService) getGenres(w rest.ResponseWriter, r *rest.Request) {
	search := models.Search{}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := service.dataStore.GetGenres(search.Genre, page, per_page)
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "genres not found", 400)
	}
}

func (service RestService) getGenre(w rest.ResponseWriter, r *rest.Request) {
	genreID, _ := strconv.Atoi(r.PathParam("id"))

	result, err := service.dataStore.GetGenre(uint(genreID))
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "genre not found", 400)
	}
}

func (service RestService) getSeries(w rest.ResponseWriter, r *rest.Request) {
	search := models.Search{}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := service.dataStore.GetSeries(search.Series, page, per_page)
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "series not found", 400)
	}
}

func (service RestService) getSerie(w rest.ResponseWriter, r *rest.Request) {
	serieID, _ := strconv.Atoi(r.PathParam("id"))

	result, err := service.dataStore.GetSerie(uint(serieID))
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "serie not found", 400)
	}
}

func (service RestService) getBooksByAuthorID(w rest.ResponseWriter, r *rest.Request) {
	authorID, _ := strconv.Atoi(r.PathParam("id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	no_details, _ := utils.ParseBool(r.URL.Query().Get("no-details"))
	random, _ := utils.ParseBool(r.URL.Query().Get("random"))

	log.Println("Requesting author's books ", authorID)

	result, err := service.dataStore.GetBooksByAuthorID(uint(authorID), no_details, random, page, per_page, models.Search{})
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "books not found", 400)
	}
}

func (service RestService) getBooksByGenreID(w rest.ResponseWriter, r *rest.Request) {
	genreID, _ := strconv.Atoi(r.PathParam("id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	no_details, _ := utils.ParseBool(r.URL.Query().Get("no-details"))
	random, _ := utils.ParseBool(r.URL.Query().Get("random"))

	log.Println("Requesting genre's books ", genreID)

	result, err := service.dataStore.GetBooksByGenreID(uint(genreID), no_details, random, page, per_page, models.Search{})
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "books not found", 400)
	}
}

func (service RestService) getBooksBySerieID(w rest.ResponseWriter, r *rest.Request) {
	serieID, _ := strconv.Atoi(r.PathParam("id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	per_page, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	no_details, _ := utils.ParseBool(r.URL.Query().Get("no-details"))
	random, _ := utils.ParseBool(r.URL.Query().Get("random"))

	log.Println("Requesting serie's books ", serieID)

	result, err := service.dataStore.GetBooksBySerieID(uint(serieID), no_details, random, page, per_page, models.Search{})
	if err == nil {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "books not found", 400)
	}
}

func (service RestService) downloadBook(w rest.ResponseWriter, r *rest.Request) {
	bookID, _ := strconv.Atoi(r.PathParam("id"))
	result, err := service.dataStore.GetBook(uint(bookID))
	if err == nil {
		outName := result.GetFullFilename()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+strings.Replace(url.QueryEscape(
			utils.ReplaceUnsupported(outName)), "+", "%20", -1))

		err := books.UnzipBookToWriter(service.dataDir, result, w.(http.ResponseWriter))
		if err != nil {
			rest.Error(w, "books not found", 400)
		}
	} else {
		rest.Error(w, "books not found", 400)
	}
}

func (service RestService) getLangs(w rest.ResponseWriter, r *rest.Request) {
	log.Println("Getting languages")

	result, err := service.dataStore.GetLangs()
	if err == nil && len(result) != 0 {
		w.WriteJson(&result)
	} else {
		rest.Error(w, "langs not found", 400)
	}
}
