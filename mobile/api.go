package mobile

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var tokenPwd = ""

func SetTokenPassword(s string) {
	tokenPwd = s
}

var logger *logrus.Logger

func SetLogger(l *logrus.Logger) {
	logger = l
}

var debug = false

func SetDebug(d bool) {
	debug = d
}

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

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(writer http.ResponseWriter, data interface{}) {
	writer.Header().Add("Content-Type", "application/json; charset=utf-8")
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	json.NewEncoder(writer).Encode(data)
}

func NewRestApi(port, certSRT, certKEY string, isTSL bool) {

	r := mux.NewRouter()

	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	r.HandleFunc("/api/user/login", login).Methods(http.MethodPost)
	r.HandleFunc("/api/user/balance", balance).Methods(http.MethodGet)
	r.HandleFunc("/api/user/char-turnover", charTurnover).Methods(http.MethodGet)
	r.HandleFunc("/api/user/top5", creditsTop5).Methods(http.MethodGet)
	r.HandleFunc("/api/user/debits", debits).Methods(http.MethodGet)
	r.HandleFunc("/api/user/debit-types", debitTypes).Methods(http.MethodGet)
	r.HandleFunc("/api/user/credits", credits).Methods(http.MethodGet)
	r.HandleFunc("/api/user/credit-types", creditTypes).Methods(http.MethodGet)

	r.Use(JwtAuth)

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
