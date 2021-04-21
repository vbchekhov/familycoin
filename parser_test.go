package main

import (
	"testing"
)

func Test_textToDebitCreditData(t *testing.T) {

	arr := []string{
		"1safasd",
		"100 $",
		"100 руб, зарплата",
		"101 долларов, зарплата",
		"102 евро, зарплата",
		"103 €, зарплата",
		"104 рублей, зарплата",
		"105 $, зарплата",
		"105, зарплата",
		"107 доллар, зарплата",
		"108 дол, зарплата",
		"0.03635819 эфир",
		"2603.403 собак",
	}

	for i := 0; i < len(arr); i++ {
		data, err := TextToDebitCreditData(arr[i])
		t.Logf("%s == "+data.Currency.Formatting+" %s %s, %v\n", arr[i], data.Sum, data.Currency.Name, data.Comment, err)
	}
}

func TestGenRegexp(t *testing.T) {

	// text := `^(\d{0,})(?:\s*(руб(?:лей|)|дол(?:лар|)(?:ов|)|евро|€|\$|)|)(?:,\s*(.*)|)$`
	text := `^(\d{0,})(?:\s*(%s)|)(?:,\s*(.*)|)$`
	sin := ""
	for s, _ := range currencysSynonym {
		if s == "$" {
			sin += "\\" + s + "|"
		} else {
			sin += s + "|"
		}
	}

	t.Logf(text, sin)

}
