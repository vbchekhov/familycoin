package main

import (
	"familycoin/models"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
)

// currencyRates Read current rates
// only USD and EUR
func currencyRates(c *skeleton.Context) bool {

	go models.RateUpdater()

	text := "📈 Курсы валют на сейчас:\n---\n"
	for _, currency := range models.CurrencyStorage {
		if currency.Default {
			continue
		}
		text += fmt.Sprintf("%s - %s руб.\n", currency.Name, currency.FormatFunc(currency.LastRate))
	}

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), text))

	return true
}
