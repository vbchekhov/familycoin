package models

import (
	"time"
)

type StackedChar struct {
	Categories []string `json:"categories"`
	Series     []Series `json:"series"`
}

type Series struct {
	Name  string    `json:"name"`
	Data  []float64 `json:"data"`
	Stack string    `json:"stack"`
}

func CreditMonthChar(chatId int64) *StackedChar {

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

func DebitMonthChar(chatId int64) *StackedChar {

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

type DtCtLineChar []struct {
	Date   string  `gorm:"column:date"`
	Debit  float64 `gorm:"column:debit"`
	Credit float64 `gorm:"column:credit"`
}

func (g DtCtLineChar) Convert() [][][]interface{} {

	arr := make([][][]interface{}, 2, 2)
	arr[0] = make([][]interface{}, len(g), len(g))
	arr[1] = make([][]interface{}, len(g), len(g))
	for i := range g {
		t, _ := time.Parse("2006-01-02", g[i].Date)
		arr[0][i] = []interface{}{t.Unix() * 1000, g[i].Debit}
		arr[1][i] = []interface{}{t.Unix() * 1000, g[i].Credit}
	}

	return arr
}

func DebitCreditLineChar(chatId int64) *DtCtLineChar {

	var res = DtCtLineChar{}

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	r := db.Raw(`
	select t1.date, floor(t1.credit) credit, floor(t2.debit) debit 
	from (select date_format(c.created_at, '%Y-%m-01') date, sum(c.sum * cr.last_rate) credit
		  from credits c
				   left join currencies cr on cr.id = ifnull(c.currency_type_id, @defaultCurrency)
		  where c.user_id in (
			  select distinct id
			  from users
			  where users.family_id = (select users.family_id from users where telegram_id = ?)
				 or users.telegram_id = ?)
		  group by date_format(c.created_at, '%Y-%m-01')) t1
			 left join (select date_format(d.created_at, '%Y-%m-01') date, sum(d.sum * cr.last_rate) debit
						from debits d
								 left join currencies cr on cr.id = ifnull(d.currency_type_id, @defaultCurrency)
						where d.user_id in (
							select distinct id
							from users
							where users.family_id = (select users.family_id from users where telegram_id = ?)
							   or users.telegram_id = ?)
						group by date_format(d.created_at, '%Y-%m-01')) t2 on t2.date = t1.date`,
		chatId, chatId, chatId, chatId)

	r.Scan(&res)

	return &res
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
