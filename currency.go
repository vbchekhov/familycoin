package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/coinpaprika/coinpaprika-api-go-client/coinpaprika"
	"github.com/vbchekhov/gorbkrates"
	"github.com/vbchekhov/skeleton"
	"time"
)

// currencys map where key is ISO number
var currencys map[string]Currency

// currencysSynonym where key is synonym list
var currencysSynonym map[string]Currency

// RateUpdater
func RateUpdater() {

	for {

		for _, currency := range currencys {

			if currency.Default {
				// default rate
				currency.LastRate = 1
				currency.update()
				continue
			}

			// USD and EUR
			rate, err := gorbkrates.Now(currency.Number)
			if err != nil || rate == 0 {

				// maybe crypto?
				rate, err = readCoinPaprika(currency.Number)

				if err != nil || rate == 0 {
					rate = 1
				}
			}

			currency.LastRate = rate
			currency.update()
		}

		// currencys map where key is ISO number
		currencys = currencyMap()
		// currencysSynonym where key is synonym list
		currencysSynonym = currencySynonymMap()

		logger.Printf("Rates update %s", time.Now().Format(time.Kitchen))

		time.Sleep(time.Minute * 15)

	}

}

// readCoinPaprika
func readCoinPaprika(number string) (float64, error) {

	var price float64
	paprikaClient := coinpaprika.NewClient(nil)

	opts := &coinpaprika.PriceConverterOptions{
		BaseCurrencyID: number, QuoteCurrencyID: "usd-us-dollars", Amount: 1,
	}
	result, err := paprikaClient.PriceConverter.PriceConverter(opts)
	if err == nil {
		price = *result.Price * currencys["840"].LastRate
	}

	return price, err
}

// currencyRates read current rates
// only USD and EUR
func currencyRates(c *skeleton.Context) bool {

	go RateUpdater()

	text := "📈 Курсы валют на сейчас:\n---\n"
	for _, currency := range currencys {
		if currency.Default {
			continue
		}
		text += fmt.Sprintf("%s - %s руб.\n", currency.Name, FloatToHumanFormat(currency.LastRate))
	}

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), text))

	return true
}
