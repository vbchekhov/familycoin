package models

import (
	"familycoin/binance"
	"github.com/vbchekhov/gorbkrates"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
	"strings"
	"time"
)

// CurrencyStorage map where key is ISO number
var CurrencyStorage map[string]Currency

// CurrencySynonymStorage where key is synonym list
var CurrencySynonymStorage map[string]Currency

/* Currency rates */

// Currency
type Currency struct {
	gorm.Model
	Name       string                      `gorm:"column:name"`
	ShortName  string                      `gorm:"column:short_name"`
	Code       string                      `gorm:"column:code"`
	SymbolCode string                      `gorm:"column:symbol_code"`
	Number     string                      `gorm:"column:number"`
	LastRate   float64                     `gorm:"column:last_rate"`
	Default    bool                        `gorm:"column:default"`
	Synonyms   string                      `gorm:"column:synonyms"`
	Formatting string                      `gorm:"column:formatting"`
	FormatFunc func(amount float64) string `gorm:"-"`
}

func (c *Currency) read() error {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	c.FormatFunc = func(amount float64) string { return message.NewPrinter(language.Russian).Sprintf(c.Formatting, amount) }

	return nil
}
func (c *Currency) Update() error {

	res := db.Save(c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func DefaultCurrency() *Currency {
	c := &Currency{Default: true}
	c.read()
	return c
}

// Currencys
type Currencys []Currency

// FirstFilling
func (c *Currencys) FirstFilling() error {

	arr := Currencys{
		{
			Name:       "🇷🇺 RUB",
			ShortName:  "руб.",
			Code:       "643",
			SymbolCode: "₽",
			Number:     "643",
			Default:    true,
			Synonyms:   "руб,рублей,₽",
			Formatting: "%.f",
		},
		{
			Name:       "🇺🇸 USD",
			ShortName:  "usd.",
			Code:       "USD",
			SymbolCode: "$",
			Number:     "840",
			Default:    false,
			Synonyms:   "дол,доллар,долларов,$,usd",
			Formatting: "%.f",
		},
		{
			Name:       "🇪🇺 EUR",
			ShortName:  "eur.",
			Code:       "EUR",
			SymbolCode: "€",
			Number:     "978",
			Default:    false,
			Synonyms:   "евро,eur,€",
			Formatting: "%.f",
		},
	}

	return arr.Create()
}

func (c *Currencys) Create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Currencys) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	copyC := *c
	for i := range copyC {
		formatting := copyC[i].Formatting
		copyC[i].FormatFunc = func(amount float64) string {
			return message.NewPrinter(language.Russian).Sprintf(formatting, amount)
		}
	}

	c = &copyC

	return nil
}

// GetCurrencySynonymMap where key is synonym list
func GetCurrencySynonymMap() map[string]Currency {

	res := map[string]Currency{}

	c := Currencys{}
	c.read()

	for i := range c {

		for _, syn := range strings.Split(c[i].Synonyms, ",") {
			res[syn] = c[i]
		}

		if c[i].Default {
			res[""] = c[i]
		}
	}

	return res
}

// GetCurrencyMap map where key is ISO number
func GetCurrencyMap() map[string]Currency {
	c := Currencys{}
	c.read()

	m := map[string]Currency{}

	for i := range c {
		m[c[i].Number] = c[i]
	}

	return m
}

// RateUpdater
func RateUpdater() {

	for {

		for _, currency := range CurrencyStorage {

			if currency.Default {
				// default rate
				currency.LastRate = 1
				currency.Update()
				continue
			}

			// USD and EUR
			rate, err := gorbkrates.Now(currency.Number)
			if err != nil || rate == 0 {

				// maybe crypto?
				rate, err = binance.Converter(currency.Number, "RUB", 1)
				if err != nil || rate == 0 {

					rate = 1

				}
			}

			currency.LastRate = rate
			currency.Update()
		}

		CurrencyStorage = GetCurrencyMap()
		CurrencySynonymStorage = GetCurrencySynonymMap()

		logger.Printf("Rates update %s", time.Now().Format(time.Kitchen))

		time.Sleep(time.Minute * 15)

	}

}
