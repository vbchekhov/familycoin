package main

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gorm.io/gorm"
	"strings"
)

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
func (c *Currency) update() error {

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
		},
		{
			Name:       "🇺🇸 USD",
			ShortName:  "usd.",
			Code:       "USD",
			SymbolCode: "$",
			Number:     "840",
			Default:    false,
			Synonyms:   "дол,доллар,долларов,$,usd",
		},
		{
			Name:       "🇪🇺 EUR",
			ShortName:  "eur.",
			Code:       "EUR",
			SymbolCode: "€",
			Number:     "978",
			Default:    false,
			Synonyms:   "евро,eur,€",
		},
	}

	return arr.create()
}

func (c *Currencys) create() error {

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

// currencysSynonym where key is synonym list
func currencySynonymMap() map[string]Currency {

	res := map[string]Currency{}

	c := Currencys{}
	c.read()

	for i := range c {

		// c[i].FormatFunc = func (amount float64) string {return message.NewPrinter(language.Russian).Sprintf(c[i].Formatting, amount)}

		for _, syn := range strings.Split(c[i].Synonyms, ",") {
			res[syn] = c[i]
		}

		if c[i].Default {
			res[""] = c[i]
		}
	}

	return res
}

// currencys map where key is ISO number
func currencyMap() map[string]Currency {
	c := Currencys{}
	c.read()

	m := map[string]Currency{}

	for i := range c {
		// c[i].FormatFunc = func (amount float64) string {return message.NewPrinter(language.Russian).Sprintf(c[i].Formatting, amount)}
		m[c[i].Number] = c[i]
	}

	return m
}
