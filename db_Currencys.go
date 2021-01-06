package main

import (
	"gorm.io/gorm"
	"strings"
)

/* Currency rates */

// Currency
type Currency struct {
	gorm.Model
	Name       string  `gorm:"column:name"`
	ShortName  string  `gorm:"column:short_name"`
	Code       string  `gorm:"column:code"`
	SymbolCode string  `gorm:"column:symbol_code"`
	Number     string  `gorm:"column:number"`
	LastRate   float64 `gorm:"column:last_rate"`
	Default    bool    `gorm:"column:default"`
	Synonyms   string  `gorm:"column:synonyms"`
}

func (c *Currency) read() error {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Currency) update() error {

	res := db.Save(c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// Currencys
type Currencys []Currency

func (c *Currencys) FirstFilling() error {

	// 	🇷🇺 RUB,RUB,643,1,"руб,рублей,₽",,руб.,₽
	// 	🇺🇸 USD,USD,840,0,"дол,доллар,долларов,$,usd",,usd.,$
	// 	🇪🇺 EUR,EUR,978,0,"евро,eur,€",,eur.,€

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

	return nil
}
func (c *Currencys) Map() map[string]Currency {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	arr := *c
	m := map[string]Currency{}

	for i := range arr {
		m[arr[i].Number] = arr[i]
	}

	return m
}

func currencySynonymMap() map[string]Currency {

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
func currencyMap() map[string]Currency {
	c := Currencys{}
	c.read()
	return c.Map()
}
