package main

import (
	"testing"
)

func TestRate(t *testing.T) {

	c := Currencys{}
	c.read()
	//
	// for i := range c {
	// 	cr, err := gorbkrates.Now(c[i].Number)
	// 	if err != nil {
	// 		t.Errorf("Error Now() %v", err)
	// 	}
	// 	t.Logf("%s - %.2f", c[i].Name, cr)
	// }

}
