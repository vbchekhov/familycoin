package mobile

import (
	"encoding/json"
	"familycoin/models"
	"net/http"
	"time"
)

func login(writer http.ResponseWriter, request *http.Request) {

	account := &models.User{}
	err := json.NewDecoder(request.Body).Decode(account) // decode the request body into struct and failed if any error occur
	if err != nil {
		Respond(writer, Message(false, "Invalid request"))
		return
	}

	resp := chekcLogin(account.Login, account.Password)

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

func debits(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	transactions := models.Detail(&models.Debit{}, telegramId, time.Now().Add(-time.Hour*24*60), time.Now())

	Respond(writer, transactions)

}

func credits(writer http.ResponseWriter, request *http.Request) {

	telegramId := request.Context().Value("telegram_id").(int64)

	transactions := models.Detail(&models.Credit{}, telegramId, time.Now().Add(-time.Hour*24*60), time.Now())

	Respond(writer, transactions)

}
