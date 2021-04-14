package main

import (
	"embed"
	"github.com/gorilla/mux"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

// Embed the entire directory.
//go:embed templates
var indexHTML embed.FS

//go:embed static
var staticFiles embed.FS

func StartWebServer() {

	r := mux.NewRouter()

	// // http.FS can be used to create a http Filesystem
	// var staticFS = http.FS(staticFiles)
	// fs := http.FileServer(staticFS)
	//
	// // Serve static files
	// r.Handle("/static/", fs)
	//
	// r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	r.HandleFunc("/static/css/"+`{path:\S+}`, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		file, err := fs.ReadFile(staticFiles, request.URL.Path[1:])
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Write(file)
	})
	r.HandleFunc("/static/js/"+`{path:\S+}`, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		file, err := fs.ReadFile(staticFiles, request.URL.Path[1:])
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Write(file)
	})
	r.HandleFunc("/static/img/"+`{path:\S+}`, func(writer http.ResponseWriter, request *http.Request) {
		file, err := fs.ReadFile(staticFiles, request.URL.Path[1:])
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		// если это svg
		if strings.HasSuffix(request.URL.Path, "svg") {
			writer.Header().Set("Content-Type", "image/svg+xml")
		}

		writer.Write(file)
	})

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		tmpl, _ := template.ParseFS(indexHTML, "templates/index-fm.html")
		tmpl.Execute(writer, nil)

	}).Methods(http.MethodGet)

	r.HandleFunc("/debit-credit-fm", func(writer http.ResponseWriter, request *http.Request) {

		tmpl, _ := template.ParseFS(indexHTML, "templates/debit-credit-fm.html")
		tmpl.Execute(writer, nil)

	}).Methods(http.MethodGet)

	logger.Printf("start server on :8099")
	if err := http.ListenAndServe(":8099", r); err != nil {
		logger.Printf("Error start server %v", err)
	}
}
