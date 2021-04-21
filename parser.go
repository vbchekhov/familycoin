package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// ParserResult result exec parser
type ParserResult struct {
	Sum      float64
	Comment  string
	Currency Currency
}

// word regexp
var CompiledRegexp *regexp.Regexp

// GenerateRegexpBySynonyms
func GenerateRegexpBySynonyms() string {

	text := `^([+-]?([0-9]*[.])?[0-9]+)(?:\s*(%s)|)(?:,\s*(.*)|)$`
	sin := ""
	for s, _ := range currencysSynonym {
		if s == "$" {
			sin += "\\" + s + "|"
		} else {
			sin += s + "|"
		}
	}

	return fmt.Sprintf(text, sin)
}

// TextToDebitCreditData convert message to debit|credit note
func TextToDebitCreditData(text string) (ParserResult, error) {

	var res ParserResult

	find := CompiledRegexp.FindStringSubmatch(text)

	if len(find) < 2 {
		return res, errors.New("Empty message")
	}

	if find[1] == "" {
		return res, errors.New("Empty amount")
	}

	// sum, err := strconv.Atoi(find[1])
	sum, err := strconv.ParseFloat(find[1], 64)
	if err != nil {
		return res, errors.New("Dont parse int")
	}

	comment := ""
	if len(find) == 5 {
		comment = find[4]
	}

	currency := currencySynonymMap()

	res = ParserResult{
		Sum:      (sum),
		Comment:  comment,
		Currency: currency[find[3]],
	}

	return res, nil
}
