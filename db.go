package main

import (
	"database/sql"
	"fmt"
	"github.com/vbchekhov/gorbkrates"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

// opening db with start
var db, _ = openDB()

// DebitCredit basic interface
type DebitCredit interface {
	// basic table name
	BasicTable() string
	// types table name
	TypesTable() string
	// types column name
	TypeIDName() string
	// detail report
	ReportDetail(title string, chatId int64, start, end time.Time) string
	// group report
	ReportGroup(title string, chatId int64, start, end time.Time) string
}

// open db func
func openDB() (*gorm.DB, error) {

	// подключаемся...
	db, err := gorm.Open(mysql.Open(conf.DataBase.stringConn()), &gorm.Config{})

	// сообщаем об ошибке
	if err != nil {
		log.Printf("Failed to connect mysql: %v", err)
		return nil, err
	}

	db.Debug()

	db.AutoMigrate(DebitType{}, Debit{}, CreditType{}, Credit{}, User{}, Currency{})

	return db, nil
}

// migrator
// TODO check this method
func _() {

	migration := db.Migrator()

	if !migration.HasTable(&User{}) || !migration.HasTable(&Family{}) {

		migration.CreateTable(&User{})
		migration.CreateTable(&Family{})

		for i := range conf.Bot.Users {
			u := User{TelegramId: conf.Bot.Users[i]}
			u.create()
		}
	}

	if !migration.HasTable(&DebitTypes{}) || !migration.HasTable(&Debit{}) {
		var debitTypes = map[int]string{
			1: "👨‍🎨 От феодала (зп)",
			2: "🎅 По милости царя (проекты)",
			3: "🧏‍♂️За красивые глазки",
		}

		migration.CreateTable(&Debit{})
		migration.CreateTable(&DebitTypes{})

		for i, s := range debitTypes {
			dt := &DebitType{Id: i, Name: s}
			dt.create()
		}
	}

	if !migration.HasTable(&CreditType{}) || !migration.HasTable(&Credit{}) {

		var creditTypes = map[int]string{
			1:  "🥒 Полезная еда",
			2:  "🍟 Фастфуд",
			3:  "🎬 Развекухи",
			4:  "🧖🏻‍♀️Красотища",
			5:  "🏠 Дом и все вот это",
			6:  "🚕 Покатухи",
			7:  "🎁 Подарочки",
			8:  "🛠🍀 Хобба",
			9:  "🧝🏼‍♂️Мой пиздюк",
			10: "👠👔 Шмотки",
		}

		migration.CreateTable(&Credit{})
		migration.CreateTable(&CreditType{})

		for i, s := range creditTypes {
			ct := &CreditType{Id: i, Name: s}
			ct.create()
		}
	}

	if !migration.HasTable(&CreditLimit{}) {
		migration.CreateTable(&CreditLimit{})
	}
}

/* Users */

// User
type User struct {
	gorm.Model
	TelegramId int64  `gorm:"column:telegram_id"`
	FullName   string `gorm:"column:full_name"`
	FamilyId   uint   `gorm:"column:family_id"`
}

func (u *User) create() error {

	res := db.Create(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (u *User) update() error {

	res := db.Save(u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (u *User) read() error {

	res := db.Where(u).Find(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (u *User) family() ([]User, error) {
	var users []User

	res := db.Table("users").Where("family_id", u.FamilyId).Find(&users)
	return users, res.Error
}

type Users []User

func (u *Users) read() error {

	res := db.Where(u).Find(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (u *Users) list() []int64 {
	arr := []int64{}
	users := *u
	for i := range users {
		arr = append(arr, users[i].TelegramId)
	}
	return arr
}

// Family
type Family struct {
	gorm.Model
	Owner  uint   `gorm:"column:owner"`
	Active string `gorm:"column:active"`
}

func (f *Family) create() error {

	res := db.Create(&f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (f *Family) update() error {

	res := db.Save(f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (f *Family) read() error {

	res := db.Where(f).Find(&f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

/* Type working with Debits */

type Details []struct {
	Created  time.Time
	Name     string
	Comment  string
	Currency string
	Sum      int
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
       c.short_name as currency,
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
       c.short_name as currency,
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

type Receipts struct {
	Id         int    `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	Sum        int    `gorm:"column:sum"`
	SymbolCode string `gorm:"column:symbol_code"`
	Comment    string `gorm:"column:comment"`
}

// ReceiptMessage
func Receipt(dt DebitCredit, id int) *Receipts {

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

	return res
}
func (r *Receipts) messagef() string {
	return fmt.Sprintf("📝 Чек №%d\n\n"+
		"```\n"+
		"📍Cумма операции: %d %s.\n"+
		"📍Категория: %s\n"+
		"📍Комментарий: %s```",
		r.Id, r.Sum, r.SymbolCode, r.Name, r.Comment)
}

type Debit struct {
	gorm.Model
	DebitTypeId    int    `gorm:"column:debit_type_id" gorm:"association_foreignkey: id"`
	UserId         uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum            int    `gorm:"column:sum"`
	Comment        string `gorm:"column:comment"`
	CurrencyTypeId uint   `gorm:"column:currency_type_id" gorm:"association_foreignkey: id"`
}

func (d *Debit) BasicTable() string {
	return "debits"
}
func (d *Debit) TypesTable() string {
	return "debit_types"
}
func (d *Debit) TypeIDName() string {
	return "debit_type_id"
}

func (d *Debit) SumToDefaultCurrency() float64 {

	return 0
}

func (d *Debit) SumToString() string {

	return ""
}

func (d *Debit) ReportDetail(title string, chatId int64, start, end time.Time) string {

	// title
	var text = "***Приходы за " + title + "*** 📈\n\n"
	var sum int

	// get detail report
	ad := Detail(d, chatId, start, end)
	for i := 0; i < len(ad); i++ {
		// text += ad[i].Created.Format("02.01") + " " + ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. _" + ad[i].Comment + "_\n"
		text += fmt.Sprintf("%s %s: %d %s _%s_\n", ad[i].Created.Format("02.01"), ad[i].Name, ad[i].Sum, ad[i].Currency, ad[i].Comment)
		sum += ad[i].Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}
func (d *Debit) ReportGroup(title string, chatId int64, start, end time.Time) string {

	// title
	var text = "***Приходы за " + title + "*** 📈\n\n"
	var sum int

	// get detail report
	ad := Group(d, chatId, start, end)
	for i := 0; i < len(ad); i++ {
		// text += ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. \n"
		text += fmt.Sprintf("%s: %d %s\n", ad[i].Name, ad[i].Sum, ad[i].Currency)
		sum += ad[i].Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}

func (d *Debit) create() error {

	res := db.Create(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (d *Debit) read() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type DebitType struct {
	Id   int    `gorm:"column:id" gorm:"primary_key" gorm:"AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

func (d *DebitType) create() error {

	res := db.Create(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (d *DebitType) read() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type DebitTypes []DebitType

func (d *DebitTypes) read() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (d *DebitTypes) convmap() (m map[string]string) {
	m = make(map[string]string)

	d.read()

	for _, debitType := range *d {
		m[strconv.Itoa(debitType.Id)] = debitType.Name
	}

	return
}

/* Type working with Credits */

type Credit struct {
	gorm.Model
	CreditTypeId   int    `gorm:"column:credit_type_id" gorm:"association_foreignkey: id"`
	UserId         uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum            int    `gorm:"column:sum"`
	Comment        string `gorm:"column:comment"`
	CurrencyTypeId uint   `gorm:"column:currency_type_id" gorm:"association_foreignkey: id"`
	Receipt        string `gorm:"column:receipt"`
	limit          *CreditLimitsByCategory
	telegramId     int64
}

func (c *Credit) BasicTable() string {
	return "credits"
}
func (c *Credit) TypesTable() string {
	return "credit_types"
}
func (c *Credit) TypeIDName() string {
	return "credit_type_id"
}

func (c *Credit) ReportDetail(title string, chatId int64, start, end time.Time) string {

	// title
	var text = "***Расходы за " + title + "*** 📉\n\n"
	var sum int

	// get detail report
	ad := Detail(c, chatId, start, end)
	for i := 0; i < len(ad); i++ {
		text += ad[i].Created.Format("02.01") + " " + ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. _" + ad[i].Comment + "_\n"
		sum += ad[i].Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}
func (c *Credit) ReportGroup(title string, chatId int64, start, end time.Time) string {

	// title
	var text = "***Расходы за " + title + "*** 📉\n\n"
	var sum int

	// get detail report
	ad := Group(c, chatId, start, end)
	for i := 0; i < len(ad); i++ {
		text += ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. \n"
		sum += ad[i].Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

	return text
}

func (c *Credit) create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	today := c.CreatedAt
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	c.limit = creditLimits(c.telegramId,
		c.CreditTypeId,
		start, end)

	return nil
}
func (c *Credit) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type CreditType struct {
	Id   int    `gorm:"column:id" gorm:"primary_key" gorm:"AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

func (c *CreditType) create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditType) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type CreditTypes []CreditType

func (c *CreditTypes) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditTypes) convmap() (m map[string]string) {
	m = make(map[string]string)

	c.read()

	for _, creditType := range *c {
		m[strconv.Itoa(creditType.Id)] = creditType.Name
	}

	return
}

type CreditLimit struct {
	gorm.Model
	CreditTypeId int  `gorm:"column:credit_type_id" gorm:"association_foreignkey: id"`
	UserId       uint `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	FamilyId     uint `gorm:"column:family_id" gorm:"association_foreignkey: id"`
	Limit        int  `gorm:"column:limit"`
}

func (c *CreditLimit) create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditLimit) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditLimit) update() error {

	res := db.Save(c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditLimit) delete() error {

	res := db.Unscoped().Delete(c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type CreditLimitsByCategory struct {
	CategoryId uint   `gorm:"column:category_id"`
	Name       string `gorm:"column:name"`
	Sum        int    `gorm:"column:sum"`
	Limits     int    `gorm:"column:limits"`
}

func creditLimits(chatId int64, creditType int, start, end time.Time) *CreditLimitsByCategory {

	var ctbc CreditLimitsByCategory
	r := db.Raw(`
		select creditType.id               as category_id,
			   creditType.name             as name,
			   SUM(c.sum)          as sum,
			   ifnull(cl.limit, 0) as limits
		from credits as c
				 left join credit_types creditType on c.credit_type_id = creditType.id
				 left join credit_limits cl on c.credit_type_id = cl.credit_type_id
		where c.created_at >= ?
		  and c.created_at <= ?
		  and c.credit_type_id = ?
		  and c.user_id in (
			select distinct id
			from users
			where users.family_id = (select users.family_id from users where telegram_id = ?)
			   or users.telegram_id = ?)
		group by c.credit_type_id;
	`, start, end, creditType, chatId, chatId)

	r.Scan(&ctbc)

	return &ctbc
}

/* Working in balance */

func balances(chatId int64) []struct {
	Currency string
	Balance  int
	Rate     float64
} {

	var res []struct {
		Currency string
		Balance  int
		Rate     float64
	}

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

/* Currency rates */

// Currency
type Currency struct {
	gorm.Model
	Name       string  `gorm:"column:name"`
	ShortName  string  `gorm:"column:short_name"`
	Code       string  `gorm:"column:code"`
	SymbolCode string  `gorm:"column:symbol_code"`
	Number     string  `gorm:"column:number"`
	LastRate   float64 `gorm:"column:last_rate"`
	Default    bool    `gorm:"column:default"`
	Synonyms   string  `gorm:"column:synonyms"`
}

func (c *Currency) read() error {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Currency) update() error {

	res := db.Save(c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// Currencys
type Currencys []Currency

func (c *Currencys) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Currencys) Map() map[string]Currency {

	db.Exec("select @defaultCurrency := id from currencies as c where c.`default` = 1;")

	arr := *c
	m := map[string]Currency{}

	for i := range arr {
		m[arr[i].Number] = arr[i]
	}

	return m
}

func currencySynonymMap() map[string]Currency {

	res := map[string]Currency{}

	c := Currencys{}
	c.read()

	for i := range c {
		for _, syn := range strings.Split(c[i].Synonyms, ",") {
			res[syn] = c[i]
		}

		if c[i].Default {
			res[""] = c[i]
		}
	}

	return res
}
func currencyMap() map[string]Currency {
	c := Currencys{}
	c.read()
	return c.Map()
}
