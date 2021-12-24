package models

import (
	"database/sql"
	"fmt"
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
	Date   string  `gorm:"column:date" json:"date"`
	Debit  float64 `gorm:"column:debit" json:"debit"`
	Credit float64 `gorm:"column:credit" json:"credit"`
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

// PeggyBankTable
func PeggyBankTable(chatId int64, start time.Time) ([]PeggyBank, error) {

	var bank []PeggyBank

	u := &User{TelegramId: chatId}
	u.read()

	db.Exec("set @creditFloor := 100;")
	db.Exec("set @debitPercent = 0.25;")
	db.Exec("set @investPercent = 0.2;")
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	db.Exec(fmt.Sprintf("set @family_id = %d", u.FamilyId))
	db.Exec(fmt.Sprintf("set @telegram_id = %d", u.TelegramId))

	q := db.Raw(`

	select month(created_at) as month,
           year(created_at)  as year,
		   SUM(bank_c)       as credit_bank,
		   SUM(bank_d)       as debit_bank,
		   floor((SUM(bank_c) + SUM(bank_d)) * @investPercent) as invest_bank
	from (select id,
				 created_at,
				 (CASE
					  WHEN FLOOR(CEILING(sum / @creditFloor) * @creditFloor) = sum THEN FLOOR(CEILING(sum / @creditFloor) * @creditFloor) + @creditFloor
					  ELSE FLOOR(CEILING(sum / @creditFloor) * @creditFloor)
					 END) - sum AS bank_c,
				 0              AS bank_d
		  from credits
		  where IFNULL(currency_type_id, @defaultCurrency) = @defaultCurrency
			and created_at >= @start
			and user_id in (
						select distinct id
						from users
						where users.family_id = @family_id or users.telegram_id = @telegram_id
					)
		  union all
	
		  select id,
				 created_at,
				 0,
				 FLOOR(sum * @debitPercent)
		  from debits
		  where IFNULL(currency_type_id, @defaultCurrency) = @defaultCurrency
			and created_at >= @start
			and user_id in (
						select distinct id
						from users
						where users.family_id = @family_id or users.telegram_id = @telegram_id
					)
		 ) v
	group by month(created_at), year(created_at)
	;`,
		sql.Named("start", start),
		sql.Named("family_id", u.FamilyId),
		sql.Named("telegram_id", u.TelegramId),
	).Scan(&bank)

	return bank, q.Error
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
