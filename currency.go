package main

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/gorbkrates"
	"github.com/vbchekhov/skeleton"
)

var currencys = currencyMap()
var currencysSynonym = currencySynonymMap()

func currencyRates(c *skeleton.Context) bool {

	usd, _ := gorbkrates.Now("840")
	eur, _ := gorbkrates.Now("978")

	text := "📈 Курсы валют на сейчас:\n---\n" +
		fmt.Sprintf("💶 EUR - %s руб.\n", floatToHumanFormat(eur)) +
		fmt.Sprintf("💵 USD - %s руб.", floatToHumanFormat(usd))

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), text))

	return true
}

func convert(c *skeleton.Context) bool {

	curr, _ := gorbkrates.Now(currencysSynonym[c.RegexpResult[2]].Number)
	num, _ := strconv.ParseFloat(c.RegexpResult[1], 64)

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(),
		fmt.Sprintf("По текущему курсу ≈ %s руб.", floatToHumanFormat(num*curr))))

	return true
}
