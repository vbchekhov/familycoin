package main

import (
	"embed"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

// Embed the entire directory.
//go:embed templates
var templatesHTML embed.FS

//go:embed static
var staticFiles embed.FS

type PageData struct {
	Balances []string
	Tops     []top
}

type top struct {
	// https://bulma.io/images/placeholders/128x128.png
	UserPic    string
	UserName   string
	Categories []string
}

func StartWebServer() {

	indexPage, _ := template.ParseFS(templatesHTML, "templates/index-fm.html")
	debitCreditPage, _ := template.ParseFS(templatesHTML, "templates/debit-credit-fm.html")

	r := mux.NewRouter()

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

		date := &PageData{
			Balances: []string{},
			Tops:     []top{},
		}

		getBalance := GetBalance(256674624)
		for _, b := range getBalance {
			date.Balances = append(date.Balances, fmt.Sprintf("%s - %s %s", currencys[b.Currency].Name, FloatToHumanFormat(b.Balance), currencys[b.Currency].SymbolCode))
		}

		users := User{TelegramId: 256674624}
		users.read()

		family, _ := users.Family()

		for i1, user := range family {

			date.Tops = append(date.Tops, top{
				UserName:   user.FullName,
				UserPic:    "https://bulma.io/images/placeholders/128x128.png",
				Categories: []string{},
			})

			credits := new(Credit)
			details := Top(credits, user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
			for i := range details {
				if i >= 3 {
					break
				}

				date.Tops[i1].Categories = append(
					date.Tops[i1].Categories,
					fmt.Sprintf("%s: %.f%s", details[i].Name, details[i].Sum, details[i].Currency),
				)
			}

		}

		indexPage.Execute(writer, date)

	}).Methods(http.MethodGet)

	r.HandleFunc("/debit", func(writer http.ResponseWriter, request *http.Request) {

		debitCreditPage.Execute(writer, nil)

	}).Methods(http.MethodGet)

	r.HandleFunc("/credit", func(writer http.ResponseWriter, request *http.Request) {

		debitCreditPage.Execute(writer, nil)

	}).Methods(http.MethodGet)

	logger.Printf("Start web server on :8099")
	if err := http.ListenAndServe(":8099", r); err != nil {
		logger.Printf("Error start server %v", err)
	}
}
