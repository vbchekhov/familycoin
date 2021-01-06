package main

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/gorbkrates"
	"github.com/vbchekhov/skeleton"
)

// currencys map where key is ISO number
var currencys = currencyMap()

// currencysSynonym where key is synonym list
var currencysSynonym = currencySynonymMap()

// currencyRates read current rates
// only USD and EUR
func currencyRates(c *skeleton.Context) bool {

	usd, _ := gorbkrates.Now("840")
	eur, _ := gorbkrates.Now("978")

	text := "📈 Курсы валют на сейчас:\n---\n" +
		fmt.Sprintf("💶 EUR - %s руб.\n", FloatToHumanFormat(eur)) +
		fmt.Sprintf("💵 USD - %s руб.", FloatToHumanFormat(usd))

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), text))

	return true
}

// convert command bot
// example /convert 100 usd
func convert(c *skeleton.Context) bool {

	curr, _ := gorbkrates.Now(currencysSynonym[c.RegexpResult[2]].Number)
	num, _ := strconv.ParseFloat(c.RegexpResult[1], 64)

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(),
		fmt.Sprintf("По текущему курсу ≈ %s руб.", FloatToHumanFormat(num*curr))))

	return true
}
