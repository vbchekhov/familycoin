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
	}

	for i := 0; i < len(arr); i++ {
		data, err := TextToDebitCreditData(arr[i])
		t.Logf("%s\t%+v, %v\n", arr[i], data, err)
	}
}
