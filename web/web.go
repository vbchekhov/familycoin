package web

import (
	"embed"
	"familycoin/models"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var logger *logrus.Logger

func SetLogger(l *logrus.Logger) {
	logger = l
}

var debug = false

func SetDebug(d bool) {
	debug = d
}

// Embed the entire directory.
//go:embed templates
var templatesHTML embed.FS

//go:embed static
var staticFiles embed.FS

// monthf russian name
func monthf(mounth time.Month) string {

	months := map[time.Month]string{
		time.January:   "❄️ Январь",
		time.February:  "🌨 Февраль",
		time.March:     "💃 Март",
		time.April:     "🌸 Апрель",
		time.May:       "🕊 Май",
		time.June:      "🌞 Июнь",
		time.July:      "🍉 Июль",
		time.August:    "⛱ Август",
		time.September: "🍁 Сентябрь",
		time.October:   "🍂 Октябрь",
		time.November:  "🥶 Ноябрь",
		time.December:  "🎅 Декабрь",
	}

	return months[mounth]
}

type PageData struct {
	Title     string
	Type      string
	BotName   string
	User      *models.User
	PeggyBank []models.PeggyBank
	Balances  []string
	Tops      []top
	Tags      []tag
	Footer    struct {
		In, Out, Balance float64
	}
	Full           map[*year]map[*mount]map[*category]detail
	Week           models.Details
	Mount          models.Details
	TotalForTables map[string]float64
}
type (
	tag struct {
		Style    string
		Name     string
		Sum      float64
		Currency string
	}
	top struct {
		UserPic    string
		UserName   string
		Categories []category
	}
	year struct {
		Date int
		Sum  float64
	}
	mount struct {
		Date time.Month
		Sum  float64
	}
	category struct {
		Name     string
		Sum      float64
		Currency string
	}
	detail []struct {
		Id       uint
		Created  time.Time
		Name     string
		Comment  string
		Currency string
		Sum      float64
	}
)

func UpdateIndexData(data *PageData) {

	getBalance := models.GetBalance(data.User.TelegramId)
	for _, b := range getBalance {
		data.Balances = append(data.Balances, fmt.Sprintf("%s - %s %s",
			models.CurrencyStorage[b.Currency].Name,
			models.CurrencyStorage[b.Currency].FormatFunc(b.Balance),
			models.CurrencyStorage[b.Currency].SymbolCode))

		if b.Rate > 1 {
			data.Footer.Balance += b.Balance * b.Rate
		} else {
			data.Footer.Balance += b.Balance
		}
	}

	family, _ := data.User.Family()

	for i1, user := range family {

		pic := "https://bulma.io/images/placeholders/128x128.png"
		if user.UserPic != "" {
			pic = user.UserPic
		}

		data.Tops = append(data.Tops, top{
			UserName:   user.FullName,
			UserPic:    pic,
			Categories: []category{},
		})

		credits := new(models.Credit)
		details := models.Top(credits, user.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
		for i := range details {
			if i >= 3 {
				break
			}

			data.Tops[i1].Categories = append(
				data.Tops[i1].Categories,
				category{details[i].Name, details[i].Sum, details[i].Currency},
			)
		}

	}

	// 3 weeks
	for i := 0; i <= 2; i++ {
		year, week := time.Now().Add(-time.Hour * 24 * 7 * time.Duration(i)).ISOWeek()
		bank, _ := models.GetPeggyBank(data.User.TelegramId, week, year)
		data.PeggyBank = append(data.PeggyBank, bank)
	}

	debits := new(models.Debit)
	groupDebits := models.Group(debits, data.User.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
	for i := range groupDebits {
		data.Tags = append(data.Tags, tag{
			Style:    "is-primary",
			Name:     groupDebits[i].Name,
			Sum:      groupDebits[i].Sum,
			Currency: groupDebits[i].Currency,
		})

		data.Footer.In += groupDebits[i].Sum
	}

	credits := new(models.Credit)
	groupCredits := models.Group(credits, data.User.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
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
func UpdateDebitCreditData(dt models.DebitCredit, data *PageData) {

	data.Title = dt.Title()
	data.Type = dt.BasicTable()

	weeks := models.Detail(dt, data.User.TelegramId, time.Now().Add(-time.Hour*24*7), time.Now())
	data.Week = weeks

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	mounts := models.Group(dt, data.User.TelegramId, start, end)
	data.Mount = mounts

	data.Full = map[*year]map[*mount]map[*category]detail{}

	full := models.Detail(dt, data.User.TelegramId, time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local), today)

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

		y.Sum += full[i].Sum * models.CurrencySynonymStorage[full[i].Currency].LastRate
		m.Sum += full[i].Sum * models.CurrencySynonymStorage[full[i].Currency].LastRate
		cache[full[i].Name].Sum += full[i].Sum * models.CurrencySynonymStorage[full[i].Currency].LastRate
	}

}

func StartWebServer(port, certSRT, certKEY string, isTSL bool) {

	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))
	r.HandleFunc("/singin", singin).Methods(http.MethodGet)

	r.NotFoundHandler = http.HandlerFunc(notFound)

	r.HandleFunc("/", home).Methods(http.MethodGet)
	r.HandleFunc("/{types:(?:debit|credit)}", debitCredit).Methods(http.MethodGet)
	r.HandleFunc("/receipt", receipt).Methods(http.MethodPost)
	r.HandleFunc("/profile", profile).Methods(http.MethodGet)
	r.HandleFunc("/update-pwd", updatePassword).Methods(http.MethodPost)
	r.HandleFunc("/login", login).Methods(http.MethodGet)
	r.HandleFunc("/statistic", statistic).Methods(http.MethodGet)
	r.HandleFunc("/get-credit-char.json", getCreditChar).Methods(http.MethodGet)
	r.HandleFunc("/get-debit-char.json", getDebitChar).Methods(http.MethodGet)
	r.HandleFunc("/get-debit-credit-line-char.json", getDebitCreditLineChar).Methods(http.MethodGet)

	r.Use(auth)

	logger.Printf("Start web server on %s...", port)

	var errStartWebServer error
	if isTSL {
		errStartWebServer = http.ListenAndServeTLS(port, certSRT, certKEY, r)
	} else {
		errStartWebServer = http.ListenAndServe(port, r)
	}

	if errStartWebServer != nil {
		logger.Errorf("Error start web server: %v", errStartWebServer)
	}

}
