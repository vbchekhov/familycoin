package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"strconv"
	"time"
)

// opening db with start
var db, _ = openDB()

// open db func
func openDB() (*gorm.DB, error) {

	// подключаемся...
	db, err := gorm.Open("mysql", conf.DataBase.stringConn())

	// сообщаем об ошибке
	if err != nil {
		log.Printf("Failed to connect mysql: %v", err)
		return nil, err
	}

	db.Debug()

	db.AutoMigrate(DebitType{}, Debit{}, CreditType{}, Credit{}, User{})

	return db, nil
}

/* Users */

// User
type User struct {
	gorm.Model
	TelegramId int64  `gorm:"column:telegram_id"`
	FullName   string `gorm:"column:full_name"`
	FamilyId   uint   `gorm:"column:family_id"`
}

// check exist user in base
func userExist(telegramId int64) bool {

	u := User{TelegramId: telegramId}
	u.read()

	return u.ID != 0
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

/* Type working with Debits */

type Debit struct {
	gorm.Model
	DebitTypeId int    `gorm:"column:debit_type_id" gorm:"association_foreignkey: id"`
	UserId      uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum         int    `gorm:"column:sum"`
	Comment     string `gorm:"column:comment"`
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

type DebitsForTime []struct {
	Created time.Time
	Comment string
	Name    string
	Sum     int
}

func debitsForTime(startTime, endTime time.Time) DebitsForTime {

	var res = DebitsForTime{}
	r := db.Raw(`
	select
       dt.name as name,
       SUM(sum) as sum
	from debits
         left join debit_types dt on debits.debit_type_id = dt.id
	where 
		created_at >= ? and created_at <= ?
	group by
		debit_type_id;

	`, startTime, endTime)

	r.Scan(&res)

	return res
}
func debitForLastWeek() DebitsForTime {

	var debits DebitsForTime

	r := db.Raw(`
	select
       debits.created_at as created,
       dt.name as name,
       debits.comment as comment,
       debits.sum as sum
	from debits
         left join debit_types dt on debits.debit_type_id = dt.id
	where 
		created_at >= ? and created_at <= ?;`,
		time.Now().Add(-time.Hour*24*7), time.Now())

	r.Scan(&debits)

	return debits
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
}

func (c *Credit) create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Credit) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type CreditsForTime []struct {
	Created time.Time
	Name    string
	Comment string
	Sum     int
}

func creditsForTime(startTime, endTime time.Time) CreditsForTime {

	var res = CreditsForTime{}
	r := db.Raw(`
	select
	   ct.name as name,
	   SUM(sum) as sum
	from credits
		 left join credit_types ct on credits.credit_type_id = ct.id
	where 
		created_at >= ? and created_at <= ?
	group by
		credit_type_id;

	`, startTime, endTime)

	r.Scan(&res)

	return res
}
func creditForLastWeek() CreditsForTime {

	var credits CreditsForTime

	r := db.Raw(`
	select
       credits.created_at as created,
       ct.name as name,
       credits.comment as comment,
       credits.sum as sum
	from credits
         left join credit_types ct on credits.credit_type_id = ct.id
	where 
		created_at >= ? and created_at <= ?;`,
		time.Now().Add(-time.Hour*24*7), time.Now())

	r.Scan(&credits)

	return credits
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

/* Working in balance */

func currentBalance() int {
	var bal int

	t1, t2 := time.Now().Add(-time.Hour*24*365*10), time.Now()

	ad := debitsForTime(t1, t2)
	for _, s := range ad {
		bal += s.Sum
	}

	ac := creditsForTime(t1, t2)
	for _, s := range ac {
		bal -= s.Sum
	}

	return bal
}
