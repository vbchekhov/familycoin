package main

import (
	"database/sql"
	"fmt"
	"github.com/vbchekhov/gorbkrates"
	"strconv"
	"time"
)

// DebitCredit basic interface
type DebitCredit interface {
	// basic table name
	BasicTable() string
	// types table name
	TypesTable() string
	// types column name
	TypeIDName() string
	// report title
	ReportTitle(title string) string
	// detail report
	ReportDetail(title string, chatId int64, start, end time.Time) string
	// group report
	ReportGroup(title string, chatId int64, start, end time.Time) string
	// Receipts
	Receipts() *Receipts
}

func Detail(dt DebitCredit, chatId int64, start, end time.Time) Details {

	// set default currency
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	var res Details

	r := db.Raw(`
	select
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
			where users.family_id = (select users.family_id from users where telegram_id = ?) or users.telegram_id = ?);`,
		start, end, chatId, chatId)

	r.Scan(&res)

	return res
}
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
		`+dt.TypeIDName()+`, c.short_name;

	`, start, end, chatId, chatId)

	r.Scan(&res)

	return res
}

type Details []struct {
	Created  time.Time
	Name     string
	Comment  string
	Currency string
	Sum      int
}

func (ad Details) Detailsf() string {

	var text string
	var sum int

	// get detail report
	for i := 0; i < len(ad); i++ {
		text += fmt.Sprintf("%s %s: %d %s _%s_\n", ad[i].Created.Format("02.01"), ad[i].Name, ad[i].Sum, ad[i].Currency, ad[i].Comment)
		sum += ad[i].Sum
	}

	// total sum todo поправить подсчет ИТОГО
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}
func (ad Details) Groupsf() string {

	// title
	var text string
	var sum int

	// get detail report
	for i := 0; i < len(ad); i++ {
		text += fmt.Sprintf("%s: %d %s\n", ad[i].Name, ad[i].Sum, ad[i].Currency)
		sum += ad[i].Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}

type Receipts struct {
	Id         int    `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	Sum        int    `gorm:"column:sum"`
	SymbolCode string `gorm:"column:symbol_code"`
	Comment    string `gorm:"column:comment"`
	table      string
}

// ReceiptMessage
func Receipt(dt DebitCredit, id uint) *Receipts {

	res := &Receipts{}
	db.Raw(`
		select
		   d.id,
		   dt.name,
		   d.sum,
		   d.comment,
		   cr.symbol_code
		from `+dt.BasicTable()+` as d
			left join `+dt.TypesTable()+` dt on d.`+dt.TypeIDName()+` = dt.id
			left join currencies cr on d.currency_type_id = cr.id
		where d.id = ?
	`, id).Scan(&res)

	res.table = dt.BasicTable()
	return res
}

// operation ID
func (r *Receipts) OperationID() string {
	return "receipt_" + r.table + "_" + strconv.Itoa(r.Id)
}

// formatter
func (r *Receipts) Fullf() string {
	return fmt.Sprintf("📝 Чек №%d\n\n"+
		"```\n"+
		"📍Cумма операции: %d %s.\n"+
		"📍Категория: %s\n"+
		"📍Комментарий: %s```",
		r.Id, r.Sum, r.SymbolCode, r.Name, r.Comment)
}
func (r *Receipts) Shortf() string {
	return fmt.Sprintf("Убыло %d %s", r.Sum, r.SymbolCode)
}

/* Working in balance */

type Balance []struct {
	Currency string
	Balance  int
	Rate     float64
}

func balances(chatId int64) Balance {

	var res Balance

	u := &User{TelegramId: chatId}
	u.read()

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	r := db.Exec(`
	create or replace table balance (
    select
        c.number as currency,
        sum(d.sum) as debit
    from debits as d
             left join debit_types dt on d.debit_type_id = dt.id
             left join users u on u.id = d.user_id
             left join currencies c on c.id = ifnull(d.currency_type_id, @defaultCurrency)
    where d.user_id in (
        select distinct id
        from users
        where users.family_id = @family_id or users.telegram_id = @telegram_id)
    group by c.number

    union all

    select
        cr.number as currency,
        sum(-c.sum) as debit
    from credits as c
             left join credit_types ct on c.credit_type_id = ct.id
             left join users u on u.id = c.user_id
             left join currencies cr on cr.id = ifnull(c.currency_type_id, @defaultCurrency)
    where c.user_id in (
        select distinct id
        from users
        where users.family_id = @family_id or users.telegram_id = @telegram_id)
    group by cr.number
	);
	`, sql.Named("family_id", u.FamilyId),
		sql.Named("telegram_id", u.TelegramId))

	r.Raw(`select currency as currency, sum(debit) as balance from balance group by currency;`).Scan(&res)

	db.Exec(`drop table balance;`)

	for i := range res {
		res[i].Rate, _ = gorbkrates.Now(res[i].Currency)
	}

	return res
}

func (bs Balance) Balancef() string {

	var sum float64

	text := "🤴 В казне сейчас, милорд!\n"
	for _, b := range bs {
		text += fmt.Sprintf("%s - %s", currencys[b.Currency].Name, floatToHumanFormat(float64(b.Balance)))
		if b.Rate > 0 {
			text += fmt.Sprintf(" (%s в руб.)", floatToHumanFormat(float64(b.Balance)*b.Rate))
			sum += float64(b.Balance) * b.Rate
		} else {
			sum += float64(b.Balance)
		}
		text += "\n"
	}

	text += fmt.Sprintf("---\nИтого в рублях %s", floatToHumanFormat(sum))

	return text
}

/* Excel reports */

// ExcelData
type ExcelData []struct {
	Date      string `gorm:"column:date"`
	DebitCat  string `gorm:"column:debit_cat"`
	CreditCat string `gorm:"column:credit_cat"`
	DebitSum  int    `gorm:"column:debit_sum"`
	CreditSum int    `gorm:"column:credit_sum"`
	Currency  string `gorm:"column:currency"`
	Comment   string `gorm:"column:comment"`
	UserName  string `gorm:"column:user_name"`
}

// read
func (e *ExcelData) read(u *User) error {

	// select date_format(d.created_at, '%d.%m.%Y') as date,
	// 	   dt.name      as debit_cat,
	// 	   ''           as credit_cat,
	// 	   d.sum        as debit_sum,
	// 	   0            as credit_sum,
	// 	   ifnull(d.comment, '') as comment,
	// 	   ifnull(u.full_name, u.telegram_id) as user_name
	//
	// from debits as d
	// 		 left join debit_types dt on d.debit_type_id = dt.id
	// 		 left join users u on u.id = d.user_id
	// where d.user_id in (
	// 	select distinct id
	// 	from users
	// 	where users.family_id = @family_id or users.telegram_id = @telegram_id)
	//
	// union all
	//
	// select date_format(c.created_at, '%d.%m.%Y') as date,
	// 	   ''           as debit_cat,
	// 	   ct.name      as credit_cat,
	// 	   0            as debit_sum,
	// 	   c.sum        as credit_sum,
	// 	   ifnull(c.comment, '') as comment,
	// 	   ifnull(u.full_name, u.telegram_id) as user_name
	//
	// from credits as c
	// 		 left join credit_types ct on c.credit_type_id = ct.id
	// 		 left join users u on u.id = c.user_id
	// where c.user_id in (
	// 	select distinct id
	// 	from users
	// 	where users.family_id = @family_id or users.telegram_id = @telegram_id)
	//
	// order by date asc

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	res := db.Raw(`
	select date_format(d.created_at, '%d.%m.%Y') as date,
			   dt.name   as debit_cat,
			   ''           		as credit_cat,
			   d.sum       as debit_sum,
			   cr.symbol_code			as currency,
			   0            	as credit_sum,
			   ifnull(d.comment, '') as comment,
			   ifnull(u.full_name, u.telegram_id) as user_name
		
		from debits as d
				 left join debit_types dt on d.debit_type_id = dt.id
				 left join users u on u.id = d.user_id
				 left join currencies cr on cr.id = ifnull(d.currency_type_id, @defaultCurrency)
		where d.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
		
		union all
		
		select date_format(c.created_at, '%d.%m.%Y') as date,
			   ''           as debit_cat,
			   ct.name      as credit_cat,
			   0            as debit_sum,
			   cr.symbol_code			as currency,
			   c.sum        as credit_sum,
			   ifnull(c.comment, '') as comment,
			   ifnull(u.full_name, u.telegram_id) as user_name
		
		from credits as c
				 left join credit_types ct on c.credit_type_id = ct.id
				 left join users u on u.id = c.user_id
				 left join currencies cr on cr.id = ifnull(c.currency_type_id, @defaultCurrency)
		where c.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
		
		order by date asc

	`,
		sql.Named("family_id", u.FamilyId),
		sql.Named("telegram_id", u.TelegramId))

	res.Scan(&e)

	if res.Error != nil {
		return res.Error
	}

	return nil
}
