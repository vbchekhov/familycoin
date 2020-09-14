package main

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strconv"
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

	db.AutoMigrate(DebitType{}, Debit{}, CreditType{}, Credit{}, User{})

	return db, nil
}

// migrator
func migrator() {

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

// array user in family
func myFamily(familyId uint) []User {

	var users []User
	u := &User{FamilyId: familyId}
	db.Table("users").Where(u).Find(&users)

	return users
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

/* Excel working methods */

type ExcelData []struct {
	Date      string `gorm:"column:date"`
	DebitCat  string `gorm:"column:debit_cat"`
	CreditCat string `gorm:"column:credit_cat"`
	DebitSum  int    `gorm:"column:debit_sum"`
	CreditSum int    `gorm:"column:credit_sum"`
	Comment   string `gorm:"column:comment"`
	UserName  string `gorm:"column:user_name"`
}

func (e *ExcelData) read(u *User) error {

	res := db.Raw(`
	select date_format(d.created_at, '%d.%m.%Y') as date,
		   dt.name      as debit_cat,
		   ''           as credit_cat,
		   d.sum        as debit_sum,
		   0            as credit_sum,
		   ifnull(d.comment, '') as comment,
		   ifnull(u.full_name, u.telegram_id) as user_name
	
	from debits as d
			 left join debit_types dt on d.debit_type_id = dt.id
			 left join users u on u.id = d.user_id
	where d.user_id in (
		select distinct id
		from users
		where users.family_id = @family_id or users.telegram_id = @telegram_id)
	
	union all
	
	select date_format(c.created_at, '%d.%m.%Y') as date,
		   ''           as debit_cat,
		   ct.name      as credit_cat,
		   0            as debit_sum,
		   c.sum        as credit_sum,
		   ifnull(c.comment, '') as comment,
		   ifnull(u.full_name, u.telegram_id) as user_name
	
	from credits as c
			 left join credit_types ct on c.credit_type_id = ct.id
			 left join users u on u.id = c.user_id
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

/* Type working with Debits */

type Details []struct {
	Created time.Time
	Name    string
	Comment string
	Sum     int
}

func Group(dt DebitCredit, chatId int64, start, end time.Time) Details {

	var res = Details{}
	r := db.Raw(`
	select
       dt.name as name,
       SUM(sum) as sum
	from `+dt.BasicTable()+`
         left join `+dt.TypesTable()+` dt on `+dt.BasicTable()+`.`+dt.TypeIDName()+` = dt.id
	where 
		created_at >= ? and created_at <= ?
		and `+dt.BasicTable()+`.user_id in (
			select distinct id 
			from users 
			where users.family_id = (select users.family_id from users where telegram_id = ?) or users.telegram_id = ?)
	group by
		`+dt.TypeIDName()+`;

	`, start, end, chatId, chatId)

	r.Scan(&res)

	return res
}
func Detail(dt DebitCredit, chatId int64, start, end time.Time) Details {

	var res Details

	r := db.Raw(`
	select
       created_at as created,
       dt.name as name,
       comment as comment,
       sum as sum
	from `+dt.BasicTable()+`
         left join `+dt.TypesTable()+` dt on `+dt.BasicTable()+`.`+dt.TypeIDName()+` = dt.id
	where 
		created_at >= ? and created_at <= ?
		and `+dt.BasicTable()+`.user_id in (
			select distinct id 
			from users 
			where users.family_id = (select users.family_id from users where telegram_id = ?) or users.telegram_id = ?);`,
		start, end, chatId, chatId)

	r.Scan(&res)

	return res
}

type Receipts struct {
	Id      int    `gorm:"column:id"`
	Name    string `gorm:"column:name"`
	Sum     int    `gorm:"column:sum"`
	Comment string `gorm:"column:comment"`
}

// ReceiptMessage
func Receipt(dt DebitCredit, id int) *Receipts {

	res := &Receipts{}
	db.Raw(`
		select
		   d.id,
		   dt.name,
		   d.sum,
		   d.comment
		from `+dt.BasicTable()+` as d
			left join `+dt.TypesTable()+` dt on d.`+dt.TypeIDName()+` = dt.id
		where d.id = ?
	`, id).Scan(&res)

	return res
}

func (r *Receipts) messagef() string {
	return fmt.Sprintf("📝 Чек №%d\n\n"+
		"```\n"+
		"📍Cумма операции: %d руб.\n"+
		"📍Категория: %s\n"+
		"📍Комментарий: %s```",
		r.Id, r.Sum, r.Name, r.Comment)
}

type Debit struct {
	gorm.Model
	DebitTypeId int    `gorm:"column:debit_type_id" gorm:"association_foreignkey: id"`
	UserId      uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum         int    `gorm:"column:sum"`
	Comment     string `gorm:"column:comment"`
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

func (d *Debit) ReportDetail(title string, chatId int64, start, end time.Time) string {

	// title
	var text = "***Приходы за " + title + "*** 📈\n\n"
	var sum int

	// get detail report
	ad := Detail(d, chatId, start, end)
	for i := 0; i < len(ad); i++ {
		text += ad[i].Created.Format("02.01") + " " + ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. _" + ad[i].Comment + "_\n"
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
		text += ad[i].Name + ": " + strconv.Itoa(ad[i].Sum) + " руб. \n"
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
	CreditTypeId int    `gorm:"column:credit_type_id" gorm:"association_foreignkey: id"`
	UserId       uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum          int    `gorm:"column:sum"`
	Comment      string `gorm:"column:comment"`
	Receipt      string `gorm:"column:receipt"`
	limit        *CreditLimitsByCategory
	telegramId   int64
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

func balances(chatId int64) int {

	var res []struct {
		Balance int `gorm:"column:b"`
	}

	u := &User{TelegramId: chatId}
	u.read()

	r := db.Exec(`
	create or replace table balance (
		select
			   sum(d.sum) as debit
		from debits as d
				 left join debit_types dt on d.debit_type_id = dt.id
				 left join users u on u.id = d.user_id
		where d.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
	
		union all
	
		select
			   sum(-c.sum) as debit
		from credits as c
				 left join credit_types ct on c.credit_type_id = ct.id
				 left join users u on u.id = c.user_id
		where c.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
	);
	`, sql.Named("family_id", u.FamilyId),
		sql.Named("telegram_id", u.TelegramId))

	r.Raw(`select sum(debit) as b from balance;`).Scan(&res)

	db.Exec(`drop table balance;`)

	return res[0].Balance
}
