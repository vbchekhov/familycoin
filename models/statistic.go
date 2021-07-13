package models

import "time"

type StackedChar struct {
	Categories []string `json:"categories"`
	Series     []Series `json:"series"`
}

type Series struct {
	Name  string    `json:"name"`
	Data  []float64 `json:"data"`
	Stack string    `json:"stack"`
}

func CreditMonthChat(chatId int64) *StackedChar {

	ct := &CreditTypes{}
	ct.read()
	char := StackedChar{}
	index := 0
	for _, t := range *ct {

		if t.Name == "💱 Обмен валют" {
			continue
		}

		char.Series = append(char.Series, Series{Name: t.Name, Data: []float64{}})

		for i := 6; i > -1; i-- {

			today := time.Now().Add(-time.Hour * 24 * 30 * time.Duration(i))
			start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
			end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

			turnover := Turnover(
				&Credit{},
				chatId,
				t.Id,
				start,
				end,
			)

			if index == 0 {
				char.Categories = append(char.Categories, monthf(start.Month())+start.Format(" 2006"))
			}
			char.Series[index].Data = append(char.Series[index].Data, turnover.Sum)
			char.Series[index].Stack = monthf(start.Month())

		}
		index++
	}

	return &char
}

func DebitMonthChat(chatId int64) *StackedChar {

	ct := &DebitTypes{}
	ct.read()
	char := StackedChar{}

	for index, t := range *ct {

		char.Series = append(char.Series, Series{Name: t.Name, Data: []float64{}})

		for i := 6; i > -1; i-- {

			today := time.Now().Add(-time.Hour * 24 * 30 * time.Duration(i))
			start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
			end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

			turnover := Turnover(
				&Debit{},
				chatId,
				t.Id,
				start,
				end,
			)

			if index == 0 {
				char.Categories = append(char.Categories, monthf(start.Month())+start.Format(" 2006"))
			}
			char.Series[index].Data = append(char.Series[index].Data, turnover.Sum)
			char.Series[index].Stack = monthf(start.Month())

		}

	}

	return &char
}

// monthf russian name
func monthf(mounth time.Month) string {

	months := map[time.Month]string{
		time.January:   "❄️ Январь",
		time.February:  "🌨 Февраль",
		time.March:     "💃 Март",
		time.April:     "🌸 Апрель",
		time.May:       "🕊 Май",
		time.June:      "🌞 Июнь",
		time.July:      "🍉 Июль",
		time.August:    "⛱ Август",
		time.September: "🍁 Сентябрь",
		time.October:   "🍂 Октябрь",
		time.November:  "🥶 Ноябрь",
		time.December:  "🎅 Декабрь",
	}

	return months[mounth]
}
