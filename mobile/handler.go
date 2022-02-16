package mobile

import (
	"encoding/json"
	"familycoin/models"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func login(writer http.ResponseWriter, request *http.Request) {

	account := &models.User{}
	err := json.NewDecoder(request.Body).Decode(account) // decode the request body into struct and failed if any error occur
	if err != nil {
		Respond(writer, Message(false, "Invalid request"))
		return
	}

	resp := checkLogin(account.Login, account.Password)

	Respond(writer, resp)

}

func balance(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	now := models.GetBalance(telegramId)
	for i, b := range now {
		now[i].Currency = models.CurrencyStorage[b.Currency].Name
	}

	Respond(writer, now)

}

func charTurnover(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	arr := models.DebitCreditLineChar(telegramId)
	chart := *arr
	for i := range chart {
		parse, _ := time.Parse("2006-01-02", chart[i].Date)
		chart[i].Date = monthf(parse.Month())
	}

	Respond(writer, &chart)

}

func creditsTop5(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	credits := new(models.Credit)
	details := models.Top(credits, telegramId, time.Now().Add(-time.Hour*24*7), time.Now())

	Respond(writer, details[:])

}

func debitCategories(writer http.ResponseWriter, request *http.Request) {

	// telegramId := request.Context().Value("telegram_id").(int64)
	dt := models.GetDebitTypes()
	Respond(writer, dt)

}

func debits(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	parse, err := url.Parse(request.RequestURI)
	if err != nil {
		logger.Error(err)
		return
	}

	values := parse.Query()
	if values.Has("grouped") {
		switch values.Get("grouped") {
		case "week":
			transactions := models.GroupByDate(&models.Debit{}, telegramId, time.Now().Add(-time.Hour*24*7), time.Now())
			Respond(writer, transactions)
		case "month":
			transactions := models.GroupByDate(&models.Debit{}, telegramId, time.Now().Add(-time.Hour*24*31), time.Now())
			Respond(writer, transactions)
		case "year":
			transactions := models.GroupByDate(&models.Debit{}, telegramId, time.Now().Add(-time.Hour*24*365), time.Now())
			Respond(writer, transactions)
		}

		return
	}

	transactions := models.Detail(&models.Debit{}, telegramId, time.Now().Add(-time.Hour*24*30), time.Now())

	Respond(writer, transactions)

}

func creditCategories(writer http.ResponseWriter, request *http.Request) {

	// telegramId := request.Context().Value("telegram_id").(int64)
	ct := models.GetCreditTypes()
	Respond(writer, ct)

}

func credits(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	parse, err := url.Parse(request.RequestURI)
	if err != nil {
		logger.Error(err)
		return
	}

	values := parse.Query()
	if values.Has("grouped") {
		switch values.Get("grouped") {
		case "week":
			transactions := models.GroupByDate(&models.Credit{}, telegramId, time.Now().Add(-time.Hour*24*7), time.Now())
			Respond(writer, transactions)
		case "month":
			transactions := models.GroupByDate(&models.Credit{}, telegramId, time.Now().Add(-time.Hour*24*31), time.Now())
			Respond(writer, transactions)
		case "year":
			transactions := models.GroupByDate(&models.Credit{}, telegramId, time.Date(time.Now().Year()-1, time.January, 1, 0, 0, 0, 0, time.Local), time.Now())
			Respond(writer, transactions)
		}

		return
	}

	transactions := models.Detail(&models.Credit{}, telegramId, time.Now().Add(-time.Hour*24*60), time.Now())

	Respond(writer, transactions)

}

func receipt(writer http.ResponseWriter, request *http.Request) {

	_ = request.Context().Value("telegram_id").(int64)

	vars := mux.Vars(request)
	types := vars["types"]
	idString := vars["id"]

	id, _ := strconv.ParseUint(idString, 10, 64)

	var receipts *models.Receipts

	if types == "debit" {
		receipts = models.Receipt(&models.Debit{}, uint(id))
	} else {
		receipts = models.Receipt(&models.Credit{}, uint(id))
	}

	Respond(writer, receipts)

}
