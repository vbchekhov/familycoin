package main

import (
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

// opening db with start
var db, _ = OpenDB()

// open db func
func OpenDB() (*gorm.DB, error) {

	// подключаемся...
	db, err := gorm.Open(mysql.Open(conf.DataBase.ConnToMariaDB()), &gorm.Config{})

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
func dbMigrator() {

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

	if !migration.HasTable(&Currency{}) {
		migration.CreateTable(&Currency{})
		cs := Currencys{}
		cs.FirstFilling()
	}

	if migration.HasTable(&Currency{}) {
		cs := Currencys{}
		cs.read()
		if len(cs) == 0 {
			cs.FirstFilling()
		}
	}
}
