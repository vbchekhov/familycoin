package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Embed the entire directory.
//go:embed templates
var templatesHTML embed.FS

//go:embed static
var staticFiles embed.FS

var funcs = template.FuncMap{
	"humanF":     func(i float64) string { return message.NewPrinter(language.Russian).Sprintf("%.f", i) },
	"dateShortF": func(t time.Time) string { return t.Format("02.01") },
	"monthF":     func(m time.Month) string { return monthf(m) },
}

type PageData struct {
	user     *User
	Balances []string
	Tops     []top
	Tags     []tag
	Footer   struct {
		In, Out, Balance float64
	}
	Title string
	Type  string
	Full  map[*year]map[*mount]map[*category]detail
	Week  Details
	Mount Details
}
type tag struct {
	Style    string
	Name     string
	Sum      float64
	Currency string
}
type top struct {
	UserPic    string
	UserName   string
	Categories []string
}
type year struct {
	Date int
	Sum  float64
}
type mount struct {
	Date time.Month
	Sum  float64
}
type category struct {
	Name string
	Sum  float64
}
type detail []struct {
	Id       uint
	Created  time.Time
	Name     string
	Comment  string
	Currency string
	Sum      float64
}

func UpdateIndexData(data *PageData) {

	getBalance := GetBalance(data.user.TelegramId)
	for _, b := range getBalance {
		data.Balances = append(data.Balances, fmt.Sprintf("%s - %s %s", currencys[b.Currency].Name, FloatToHumanFormat(b.Balance), currencys[b.Currency].SymbolCode))
		if b.Rate > 0 {
			data.Footer.Balance += b.Balance * b.Rate
		} else {
			data.Footer.Balance += b.Balance
		}
	}

	family, _ := data.user.Family()

	for i1, user := range family {

		pic := "https://bulma.io/images/placeholders/128x128.png"
		if user.UserPic != "" {
			pic = user.UserPic
		}

		data.Tops = append(data.Tops, top{
			UserName:   user.FullName,
			UserPic:    pic,
			Categories: []string{},
		})

		credits := new(Credit)
		details := Top(credits, user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
		for i := range details {
			if i >= 3 {
				break
			}

			data.Tops[i1].Categories = append(
				data.Tops[i1].Categories,
				fmt.Sprintf("%s: %.f%s", details[i].Name, details[i].Sum, details[i].Currency),
			)
		}

	}

	debits := new(Debit)
	groupDebits := Group(debits, data.user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
	for i := range groupDebits {
		data.Tags = append(data.Tags, tag{
			Style:    "is-primary",
			Name:     groupDebits[i].Name,
			Sum:      groupDebits[i].Sum,
			Currency: groupDebits[i].Currency,
		})

		data.Footer.In += groupDebits[i].Sum
	}

	credits := new(Credit)
	groupCredits := Group(credits, data.user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
	for i := range groupCredits {
		data.Tags = append(data.Tags, tag{
			Style:    "is-link",
			Name:     groupCredits[i].Name,
			Sum:      groupCredits[i].Sum,
			Currency: groupCredits[i].Currency,
		})

		data.Footer.Out += groupCredits[i].Sum
	}
}
func UpdateDebitCreditData(dt DebitCredit, data *PageData) {

	data.Title = dt.Title()
	data.Type = dt.BasicTable()

	weeks := Detail(dt, data.user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
	data.Week = weeks

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	mounts := Group(dt, data.user.TelegramId, start, end)
	data.Mount = mounts

	data.Full = map[*year]map[*mount]map[*category]detail{}

	full := Detail(dt, data.user.TelegramId, time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local), today)

	y := &year{}
	m := &mount{}
	c := &category{}

	cache := map[string]*category{}

	for i := range full {

		if y.Date != full[i].Created.Year() {
			y = &year{Date: full[i].Created.Year(), Sum: 0}
			data.Full[y] = map[*mount]map[*category]detail{}
		}

		if m.Date != full[i].Created.Month() {
			m = &mount{Date: full[i].Created.Month(), Sum: 0}
			cache = map[string]*category{}
			data.Full[y][m] = map[*category]detail{}
		}

		if full[i].Name == "💱 Обмен валют" {
			continue
		}

		if _, ok := cache[full[i].Name]; !ok {
			c = &category{Name: full[i].Name, Sum: 0}
			cache[full[i].Name] = c
			data.Full[y][m][cache[full[i].Name]] = detail{}
		}

		data.Full[y][m][cache[full[i].Name]] = append(data.Full[y][m][cache[full[i].Name]], full[i])

		y.Sum += full[i].Sum
		m.Sum += full[i].Sum
		cache[full[i].Name].Sum += full[i].Sum
	}

}

func StartWebServer() {

	indexPage, _ := template.New("index.html").Funcs(funcs).ParseFS(templatesHTML, "templates/index.html")
	homePage, _ := template.New("home.html").Funcs(funcs).ParseFS(templatesHTML, "templates/home.html")
	debitCreditPage, _ := template.New("debit-credit.html").Funcs(funcs).ParseFS(templatesHTML, "templates/debit-credit.html")

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
	r.HandleFunc("/img/"+`{path:\S+}`, func(writer http.ResponseWriter, request *http.Request) {
		file, err := ioutil.ReadFile(request.URL.Path[1:])
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

		token := request.Context().Value("token").(string)
		user := sessions.Map[token]

		date := PageData{
			user:     user,
			Balances: []string{},
			Tops:     []top{},
			Tags:     []tag{},
		}

		UpdateIndexData(&date)

		err3 := homePage.Execute(writer, date)
		if err3 != nil {
			logger.Error(err3)
		}

	}).Methods(http.MethodGet)

	r.HandleFunc("/debit", func(writer http.ResponseWriter, request *http.Request) {

		token := request.Context().Value("token").(string)
		user := sessions.Map[token]

		date := PageData{
			user: user,
		}

		UpdateDebitCreditData(&Debit{}, &date)

		err := debitCreditPage.Execute(writer, date)
		if err != nil {
			logger.Error(err)
		}
	}).Methods(http.MethodGet)
	r.HandleFunc("/credit", func(writer http.ResponseWriter, request *http.Request) {

		token := request.Context().Value("token").(string)
		user := sessions.Map[token]

		date := PageData{
			user: user,
		}

		UpdateDebitCreditData(&Credit{}, &date)

		err := debitCreditPage.Execute(writer, date)
		if err != nil {
			logger.Error(err)
		}

	}).Methods(http.MethodGet)

	r.HandleFunc("/receipt", func(writer http.ResponseWriter, request *http.Request) {

		defer request.Body.Close()

		var Request struct {
			Type string `json:"type"`
			Id   uint   `json:"id"`
		}

		var receipts *Receipts

		all, _ := ioutil.ReadAll(request.Body)
		err := json.Unmarshal(all, &Request)

		log.Print(string(all), err, Request)

		if Request.Type == "debits" {
			receipts = Receipt(&Debit{}, Request.Id)
		} else {
			receipts = Receipt(&Credit{}, Request.Id)
		}

		b, _ := json.Marshal(receipts)
		writer.Write(b)

	}).Methods(http.MethodPost)

	r.HandleFunc("/singin", func(writer http.ResponseWriter, request *http.Request) {

		err := indexPage.Execute(writer, nil)
		if err != nil {
			logger.Error(err)
		}

	}).Methods(http.MethodGet)

	r.HandleFunc("/login", login).Methods(http.MethodGet)

	r.Use(auth)

	logger.Printf("Start web server on :8099")
	if err := http.ListenAndServe(":8099", r); err != nil {
		logger.Printf("Error start server %v", err)
	}
}
