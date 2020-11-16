package main

import (
	"errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"regexp"
	"strconv"
)

func textToDebitCreditData(text string) (struct {
	Sum      int
	Comment  string
	Currency Currency
}, error) {

	res := struct {
		Sum      int
		Comment  string
		Currency Currency
	}{}

	mc := regexp.MustCompile(`^(\d{0,})(?:\s*(руб(?:лей|)|дол(?:лар|)(?:ов|)|евро|€|\$|)|)(?:,\s*(.*)|)$`)
	find := mc.FindStringSubmatch(text)

	if len(find) < 2 {
		return res, errors.New("Empty message")
	}

	if find[1] == "" {
		return res, errors.New("Empty amount")
	}

	sum, err := strconv.Atoi(find[1])
	if err != nil {
		return res, errors.New("Dont parse int")
	}

	comment := ""
	if len(find) == 4 {
		comment = find[3]
	}

	currency := currencySynonymMap()

	res = struct {
		Sum      int
		Comment  string
		Currency Currency
	}{
		Sum:      sum,
		Comment:  comment,
		Currency: currency[find[2]],
	}

	return res, nil
}

func floatToHumanFormat(amount float64) string {
	return message.NewPrinter(language.Russian).Sprintf("%.2f", amount)
}
