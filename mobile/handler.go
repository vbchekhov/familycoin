package mobile

import (
	"encoding/json"
	"familycoin/models"
	"net/http"
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

	chart := models.DebitCreditLineChar(telegramId)

	Respond(writer, &chart)

}
