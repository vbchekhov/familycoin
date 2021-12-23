package web

import (
	"encoding/json"
	"familycoin/models"
	"github.com/gorilla/mux"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"
)

var funcs = template.FuncMap{
	"humanF":     func(i float64) string { return message.NewPrinter(language.Russian).Sprintf("%.f", i) },
	"currencyF":  func(i float64, currency string) string { return models.CurrencySynonymStorage[currency].FormatFunc(i) },
	"dateShortF": func(t time.Time) string { return t.Format("02.01") },
	"date":       func(t time.Time) string { return t.Format("02.01.06 15:04:05") },
	"monthF":     func(m time.Month) string { return monthf(m) },
}

var renderStorage = map[string]*template.Template{}
var render = func(name string, f template.FuncMap, patterns ...string) *template.Template {

	defined := []string{
		"templates/header.gotm",
		"templates/sw.gotm",
	}

	for i := range defined {
		patterns = append(patterns, defined[i])
	}

	if debug {
		for i := range patterns {
			patterns[i] = "web/" + patterns[i]
		}
		fs, err := template.New(name).Funcs(f).ParseFiles(patterns...)
		if err != nil {
			logger.Errorf("Error render template %v", err)
		}
		return fs
	}

	var fs *template.Template
	if _, ok := renderStorage[name]; !ok {
		var err error

		if fs, err = template.New(name).Funcs(f).ParseFS(templatesHTML, patterns...); err != nil {
			logger.Errorf("Error render template %v", err)
		}
		renderStorage[name] = fs
	}
	return renderStorage[name]
}

func singin(writer http.ResponseWriter, request *http.Request) {

	data := PageData{BotName: BotName}
	render("index.html", funcs, "templates/index.html").Execute(writer, data)
}
func home(writer http.ResponseWriter, request *http.Request) {

	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	date := PageData{
		User:     user,
		Balances: []string{},
		Tops:     []top{},
		Tags:     []tag{},
	}

	UpdateIndexData(&date)

	render("home.html", funcs, "templates/home.html").Execute(writer, date)
}
func debitCredit(writer http.ResponseWriter, request *http.Request) {

	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	vars := mux.Vars(request)
	types := vars["types"]

	date := PageData{
		User: user,
	}

	if types == "debit" {
		UpdateDebitCreditData(&models.Debit{}, &date)
	} else {
		UpdateDebitCreditData(&models.Credit{}, &date)
	}

	render("debit-credit.html", funcs, "templates/debit-credit.html").Execute(writer, date)
}
func receipt(writer http.ResponseWriter, request *http.Request) {

	defer request.Body.Close()

	var Request struct {
		Type string `json:"type"`
		Id   uint   `json:"id"`
	}

	var receipts *models.Receipts

	all, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(all, &Request)

	if Request.Type == "debits" {
		receipts = models.Receipt(&models.Debit{}, Request.Id)
	} else {
		receipts = models.Receipt(&models.Credit{}, Request.Id)
	}

	b, _ := json.Marshal(receipts)
	writer.Write(b)

}
func profile(writer http.ResponseWriter, request *http.Request) {

	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	user.Read()

	date := map[string]interface{}{
		"User": user,
	}

	render("profile.html", funcs, "templates/profile.html").Execute(writer, date)
}
func statistic(writer http.ResponseWriter, request *http.Request) {

	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	pb, err := models.PeggyBankTable(user.TelegramId, time.Date(2021, time.April, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		logger.Error(err)
	}

	date := PageData{
		Title:          "📈 Статистика",
		User:           user,
		PeggyBank:      pb,
		TotalForTables: map[string]float64{},
	}

	for i := range date.PeggyBank {
		date.TotalForTables["totalDebitBank"] += date.PeggyBank[i].DebitBank
		date.TotalForTables["totalCreditBank"] += date.PeggyBank[i].CreditBank
		date.TotalForTables["totalInvestBank"] += date.PeggyBank[i].InvestBank
	}

	logger.Error(render("statistic.html", funcs, "templates/statistic.html").Execute(writer, date))
}

func getCreditChar(writer http.ResponseWriter, request *http.Request) {
	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	json.NewEncoder(writer).Encode(models.CreditMonthChar(user.TelegramId))
}

func getDebitChar(writer http.ResponseWriter, request *http.Request) {
	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	json.NewEncoder(writer).Encode(models.DebitMonthChar(user.TelegramId))
}

func getDebitCreditLineChar(writer http.ResponseWriter, request *http.Request) {
	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	json.NewEncoder(writer).Encode(models.DebitCreditLineChar(user.TelegramId).Convert())
}

func updatePassword(writer http.ResponseWriter, request *http.Request) {

	token := request.Context().Value("token").(string)
	user := sessions.Map[token]

	defer request.Body.Close()

	body, _ := ioutil.ReadAll(request.Body)

	user.Password = string(body)
	user.Update()

	writer.Write([]byte("OK"))
}
