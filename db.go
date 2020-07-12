package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"strconv"
	"time"
)

var db, _ = openDB()

func conn() string {
	return fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True",
		"---",
		"---",
		"familycoin")
}

func openDB() (*gorm.DB, error) {

	// подключаемся...
	db, err := gorm.Open("mysql", conn())

	// сообщаем об ошибке
	if err != nil {
		log.Printf("Failed to connect mysql: %v", err)
		return nil, err
	}

	db.Debug()

	db.AutoMigrate(DebitType{}, Debit{}, CreditType{}, Credit{}, User{})

	return db, nil
}

// --- Users

type User struct {
	gorm.Model
	TelegramId int64  `gorm:"column:telegram_id"`
	FullName   string `gorm:"column:full_name"`
}

func userExist(telegramId int64) bool {

	u := User{TelegramId: telegramId}
	u.get()

	return u.ID != 0
}

func (u *User) set() error {

	res := db.Create(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (u *User) get() error {

	res := db.Where(u).Find(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// type Users []User

// func (u *Users) get() error {
//
// 	res := db.Where(u).Find(&u)
// 	if res.Error != nil {
// 		return res.Error
// 	}
//
// 	return nil
// }

// --- Users

// --- DebitType

type DebitsForTime []struct {
	Name string
	Sum  int
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

type DebitType struct {
	Id   int    `gorm:"column:id" gorm:"primary_key" gorm:"AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

func (d *DebitType) set() error {

	res := db.Create(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (d *DebitType) get() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

type DebitTypes []DebitType

func (d *DebitTypes) get() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (d *DebitTypes) convmap() (m map[string]string) {
	m = make(map[string]string)

	d.get()

	for _, debitType := range *d {
		m[strconv.Itoa(debitType.Id)] = debitType.Name
	}

	return
}

// --- DebitType

type Debit struct {
	gorm.Model
	DebitTypeId int  `gorm:"column:debit_type_id" gorm:"association_foreignkey: id"`
	UserId      uint `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum         int  `gorm:"column:sum"`
}

func (d *Debit) set() error {

	res := db.Create(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (d *Debit) get() error {

	res := db.Where(d).Find(&d)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// --- CreditType

type CreditsForTime []struct {
	Name string
	Sum  int
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

type CreditTypes []CreditType

func (c *CreditTypes) get() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (c *CreditTypes) convmap() (m map[string]string) {
	m = make(map[string]string)

	c.get()

	for _, creditType := range *c {
		m[strconv.Itoa(creditType.Id)] = creditType.Name
	}

	return
}

type CreditType struct {
	Id   int    `gorm:"column:id" gorm:"primary_key" gorm:"AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

func (c *CreditType) set() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (c *CreditType) get() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// --- CreditType

type Credit struct {
	gorm.Model
	CreditTypeId int    `gorm:"column:credit_type_id" gorm:"association_foreignkey: id"`
	UserId       uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum          int    `gorm:"column:sum"`
	Comment      string `gorm:"column:comment"`
	Receipt      string `gorm:"column:receipt"`
}

func (c *Credit) set() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (c *Credit) get() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
