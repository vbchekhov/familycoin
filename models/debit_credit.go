package models

import (
	"fmt"
	"time"
)

// DebitCredit basic interface
type DebitCredit interface {
	Title() string
	ReceiptFile() string
	// BasicTable basic table name
	BasicTable() string
	// TypesTable types table name
	TypesTable() string
	// TypeIDName types column name
	TypeIDName() string
	// ReportTitle report title
	ReportTitle(title string) string
	// ReportDetail detail report
	ReportDetail(title string, chatId int64, start, end time.Time) string
	// ReportGroup group report
	ReportGroup(title string, chatId int64, start, end time.Time) string
	// Receipts operation
	Receipts() *Receipts
}

// Detail Read detail notes for period
func Detail(dt DebitCredit, chatId int64, start, end time.Time) Details {

	// set default currency
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	var res Details

	r := db.Raw(`
	select
	   dc.id as id,
       dc.created_at as created,
       dt.name as name,
       dc.comment as comment,
       c.symbol_code as currency,
       dc.sum as sum
	from `+dt.BasicTable()+` as dc
         left join `+dt.TypesTable()+` dt on dc.`+dt.TypeIDName()+` = dt.id
		 left join currencies c on ifnull(dc.currency_type_id, @defaultCurrency) = c.id 
	where 
		dc.created_at >= ? and dc.created_at <= ?
		and dc.user_id in (
			select distinct id 
			from users 
			where users.family_id = (select users.family_id from users where telegram_id = ?) or users.telegram_id = ?)
	order by dc.created_at, dt.name asc;`,
		start, end, chatId, chatId)

	r.Scan(&res)

	return res
}

// Group Read group by type name and currency
func Group(dt DebitCredit, chatId int64, start, end time.Time) Details {

	// set default currency
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	var res = Details{}

	r := db.Raw(`
	select
       dt.name as name,
       c.symbol_code as currency,
       SUM(dc.sum) as sum
	from `+dt.BasicTable()+` as dc
         left join `+dt.TypesTable()+` dt on dc.`+dt.TypeIDName()+` = dt.id
		 left join currencies c on ifnull(dc.currency_type_id, @defaultCurrency) = c.id
	where 
		dc.created_at >= ? and dc.created_at <= ?
		and dc.user_id in (
			select distinct id 
			from users 
			where users.family_id = (select users.family_id from users where telegram_id = ?) or users.telegram_id = ?)
	group by
		`+dt.TypeIDName()+`, c.short_name
	order by sum desc;

	`, start, end, chatId, chatId)

	r.Scan(&res)

	return res
}

// Top Read group by type name and currency, order by sum
func Top(dt DebitCredit, chatId int64, start, end time.Time) Details {

	// set default currency
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	var res = Details{}

	r := db.Raw(`
	select
       dt.name as name,
       c.symbol_code as currency,
       SUM(dc.sum) as sum
	from `+dt.BasicTable()+` as dc
         left join `+dt.TypesTable()+` dt on dc.`+dt.TypeIDName()+` = dt.id
		 left join currencies c on ifnull(dc.currency_type_id, @defaultCurrency) = c.id
	where 
		dc.created_at >= ? and dc.created_at <= ?
		and dc.user_id  = (
			select users.id from users where telegram_id = ?
		)
	group by
		`+dt.TypeIDName()+`, c.short_name
	order by sum desc;

	`, start, end, chatId)

	r.Scan(&res)

	return res
}

// Details array details
type Details []struct {
	Id       uint
	Created  time.Time
	Name     string
	Comment  string
	Currency string
	Sum      float64
}

// Detailsf format message after send
func (ad Details) Detailsf() string {

	// title
	var text string
	// total total
	var total float64

	// get detail report
	for i := 0; i < len(ad); i++ {
		// check currency`s
		var rate float64 = 1
		var c, ok = Currency{}, false
		if c, ok = CurrencySynonymStorage[ad[i].Currency]; ok && c.LastRate != 0 {
			rate = c.LastRate
		}
		// Create text
		text += fmt.Sprintf("%s %s: %s %s _%s_\n", ad[i].Created.Format("02.01"), ad[i].Name, c.FormatFunc(ad[i].Sum), ad[i].Currency, ad[i].Comment)
		// Update total
		total += ad[i].Sum * rate
	}

	// total sum
	text += fmt.Sprintf("---\n_Итого:_ %s %s", DefaultCurrency().FormatFunc(total), DefaultCurrency().SymbolCode)

	return text
}

// Groupsf format message after send
func (ad Details) Groupsf() string {

	// title
	var text string
	// total total
	var total float64

	// get detail report
	for i := 0; i < len(ad); i++ {
		// check currency`s
		var rate float64 = 1
		var c, ok = Currency{}, false
		if c, ok = CurrencySynonymStorage[ad[i].Currency]; ok && c.LastRate != 0 {
			rate = c.LastRate
		}
		// Create text
		text += fmt.Sprintf("%s: %s %s\n", ad[i].Name, c.FormatFunc(ad[i].Sum), ad[i].Currency)
		// Update total
		total += ad[i].Sum * rate
	}

	// total sum
	text += fmt.Sprintf("---\n_Итого:_ %s %s", DefaultCurrency().FormatFunc(total), DefaultCurrency().SymbolCode)

	return text
}
