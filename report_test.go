package main

import (
	"fmt"
	"testing"
	"time"
)

func Test_exportExcel(t *testing.T) {

	today := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	fmt.Printf("today %s\nstart %s\nend %s\n", today.String(), start.String(), end.String())

}
