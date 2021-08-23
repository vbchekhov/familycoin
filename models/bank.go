package models

import (
	"database/sql"
	"familycoin/binance"
	"fmt"
	"strconv"
	"time"
)

/* Peggy bank */

type PeggyBank struct {
	Month          time.Month `gorm:"column:month"`
	Week           int        `gorm:"column:week"`
	CreditBank     float64    `gorm:"column:credit_bank"`
	DebitBank      float64    `gorm:"column:debit_bank"`
	InvestBank     float64    `gorm:"column:invest_bank"`
	Monday, Sunday time.Time
}

// GetPeggyBank
func GetPeggyBank(chatId int64, week, year int) (PeggyBank, error) {

	var bank PeggyBank

	u := &User{TelegramId: chatId}
	u.read()

	db.Exec("set @creditFloor := 100;")
	db.Exec("set @debitPercent = 0.25;")
	db.Exec("set @investPercent = 0.2;")
	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	db.Exec(fmt.Sprintf("set @week = %d", week))
	db.Exec(fmt.Sprintf("set @year = %d", year))
	db.Exec(fmt.Sprintf("set @family_id = %d", u.FamilyId))
	db.Exec(fmt.Sprintf("set @telegram_id = %d", u.TelegramId))

	q := db.Raw(`

	select week(created_at)	as week,
		   SUM(bank_c)      as credit_bank,
		   SUM(bank_d)      as debit_bank,
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
			and week(created_at) = @week
			and year(created_at) = @year
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
			and week(created_at) = @week
			and year(created_at) = @year
			and user_id in (
						select distinct id
						from users
						where users.family_id = @family_id or users.telegram_id = @telegram_id
					)
		 ) v
	group by week(created_at)
	;`,
		sql.Named("week", week),
		sql.Named("year", year),
		sql.Named("family_id", u.FamilyId),
		sql.Named("telegram_id", u.TelegramId),
	).Scan(&bank)

	bank.Monday, bank.Sunday = DaysOfISOWeek(year, week, time.Local)

	return bank, q.Error
}

// DaysOfISOWeek
func DaysOfISOWeek(year int, week int, timezone *time.Location) (monday time.Time, sunday time.Time) {
	monday = time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	sunday = time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := monday.ISOWeek()
	for monday.Weekday() != time.Monday { // iterate back to Monday
		monday = monday.AddDate(0, 0, -1)
		sunday = monday.AddDate(0, 0, -7)
		isoYear, isoWeek = monday.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		monday = monday.AddDate(0, 0, 1)
		sunday = monday.AddDate(0, 0, 7)
		isoYear, isoWeek = monday.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		monday = monday.AddDate(0, 0, 1)
		sunday = monday.AddDate(0, 0, 7)
		isoYear, isoWeek = monday.ISOWeek()
	}
	return monday, sunday
}

/* Receipts */

// Receipts for debit|credit notes
type Receipts struct {
	Id         int       `gorm:"column:id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	Name       string    `gorm:"column:name"`
	Sum        int       `gorm:"column:sum"`
	SymbolCode string    `gorm:"column:symbol_code"`
	Comment    string    `gorm:"column:comment"`
	Receipt    string    `gorm:"column:receipt"`
	FullName   string    `gorm:"column:full_name"`
	UserPic    string    `gorm:"column:user_pic"`
	table      string    // current table name Read
}

// ReceiptMessage
func Receipt(dt DebitCredit, id uint) *Receipts {

	res := &Receipts{}
	db.Raw(`
		select
		   d.id,
		   d.created_at,	
		   dt.name,
		   u.full_name as full_name,
		   u.user_pic as user_pic,
		   d.sum,
		   d.comment,
           `+dt.ReceiptFile()+`
		   cr.symbol_code
		from `+dt.BasicTable()+` as d
			left join `+dt.TypesTable()+` dt on d.`+dt.TypeIDName()+` = dt.id
			left join currencies cr on d.currency_type_id = cr.id
			left join users u on d.user_id = u.id
		where d.id = ?
	`, id).Scan(&res)

	res.table = dt.BasicTable()
	return res
}

// OperationID operation ID in table
func (r *Receipts) OperationID() string {
	return "receipt_" + r.table + "_" + strconv.Itoa(r.Id)
}

// Fullf full message Receipts
func (r *Receipts) Fullf() string {
	return fmt.Sprintf("📝 Чек №%d\n\n"+
		"```\n"+
		"📍Cумма операции: %d %s.\n"+
		"📍Категория: %s\n"+
		"📍Комментарий: %s```",
		r.Id, r.Sum, r.SymbolCode, r.Name, r.Comment)
}

// Shortf short message Receipts
func (r *Receipts) Shortf() string {
	t := "Прибыло"
	if r.table == "credits" {
		t = "Убыло"
	}
	return fmt.Sprintf("%s %d %s", t, r.Sum, r.SymbolCode)
}

/* Working in balance */

// Balance array group by Currency
type Balance []struct {
	Balance  float64
	Currency string
	Rate     float64
}

// GetBalance Read current balance
func GetBalance(chatId int64) Balance {

	var res Balance

	u := &User{TelegramId: chatId}
	u.read()

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	db.Raw(`
	select currency,
		sum(debit) as balance
	from (
			select c.number as currency,
				sum(d.sum) as debit
			from debits as d
				left join debit_types dt on d.debit_type_id = dt.id
				left join users u on u.id = d.user_id
				left join currencies c on c.id = ifnull(d.currency_type_id, @defaultCurrency)
			where d.user_id in (
					select distinct id
					from users
					where users.family_id = @family_id or users.telegram_id = @telegram_id
				)
			and d.sum <> 0
			group by c.number

			union all

			select cr.number as currency,
				sum(- c.sum) 
			from credits as c
				left join credit_types ct on c.credit_type_id = ct.id
				left join users u on u.id = c.user_id
				left join currencies cr on cr.id = ifnull(c.currency_type_id, @defaultCurrency)
			where c.user_id in (
					select distinct id
					from users
					where users.family_id = @family_id or users.telegram_id = @telegram_id
				)
			and c.sum <> 0
			group by cr.number
		) t
	group by currency
	`, sql.Named("family_id", u.FamilyId), sql.Named("telegram_id", u.TelegramId)).Scan(&res)

	res = append(res, binance.Balance()...)

	for i := range res {
		res[i].Rate = GetCurrencyMap()[res[i].Currency].LastRate
	}

	return res
}

// Balancef format message
func (bs Balance) Balancef() string {

	var sum float64

	text := "🤴 В казне сейчас, милорд!\n"
	for _, b := range bs {
		text += fmt.Sprintf("%s - %s", CurrencyStorage[b.Currency].Name, CurrencyStorage[b.Currency].FormatFunc(b.Balance))
		if b.Rate > 1 {
			text += fmt.Sprintf(" (%s в %s)", DefaultCurrency().FormatFunc(b.Balance*b.Rate), DefaultCurrency().ShortName)
			sum += b.Balance * b.Rate
		} else {
			sum += b.Balance
		}
		text += "\n"
	}

	text += fmt.Sprintf("---\nИтого %s %s", DefaultCurrency().FormatFunc(sum), DefaultCurrency().SymbolCode)

	return text
}

/* Excel reports */

// ExcelData
type ExcelData []struct {
	Date      string  `gorm:"column:date"`
	DebitCat  string  `gorm:"column:debit_cat"`
	CreditCat string  `gorm:"column:credit_cat"`
	DebitSum  float64 `gorm:"column:debit_sum"`
	CreditSum float64 `gorm:"column:credit_sum"`
	Currency  string  `gorm:"column:currency"`
	Comment   string  `gorm:"column:comment"`
	UserName  string  `gorm:"column:user_name"`
}

// Exchange
func (e ExcelData) Exchange() {

	for i := range e {
		if e[i].Currency != DefaultCurrency().Number {
			if c, ok := CurrencyStorage[e[i].Currency]; ok {
				if e[i].DebitSum != 0 {
					e[i].DebitSum = c.LastRate * e[i].DebitSum
				}
				if e[i].CreditSum != 0 {
					e[i].CreditSum = c.LastRate * e[i].CreditSum
				}
			}
		}
	}
}

// Read
func (e *ExcelData) Read(u *User) error {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	res := db.Raw(`
	select 
		   date_format(d.created_at, '%d.%m.%Y') as date,
		   dt.name   			 as debit_cat,
		   ''           		 as credit_cat,
		   d.sum       			 as debit_sum,
		   cr.number		 	 as currency,
		   0            		 as credit_sum,
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
	
	select 
		   date_format(c.created_at, '%d.%m.%Y') as date,
		   ''           		 as debit_cat,
		   ct.name      		 as credit_cat,
		   0            		 as debit_sum,
		   cr.number    		 as currency,
		   c.sum        		 as credit_sum,
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
