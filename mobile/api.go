package mobile

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
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

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func NewRestApi(port, certSRT, certKEY string, isTSL bool) {

	r := mux.NewRouter()

	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./img"))))

	r.HandleFunc("/api/user/login", login).Methods(http.MethodPost)
	r.HandleFunc("/api/user/balance", balance).Methods(http.MethodGet)
	r.HandleFunc("/api/user/char-turnover", charTurnover).Methods(http.MethodGet)

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
