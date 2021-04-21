package main

import (
	"testing"
)

func TestDefaultCurrency(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "rub",
			want: "643",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultCurrency(); got.Number != tt.want {
				t.Errorf("DefaultCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_currencyMap(t *testing.T) {
	for s := range currencys {
		t.Log(currencys[s].FormatFunc(10), "===", currencys[s].Formatting)
	}
}
