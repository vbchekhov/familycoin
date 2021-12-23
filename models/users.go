package models

import (
	"gorm.io/gorm"
	"time"
)

/* Users */

// User
type User struct {
	gorm.Model
	TelegramId int64  `gorm:"column:telegram_id"`
	FullName   string `gorm:"column:full_name"`
	FamilyId   uint   `gorm:"column:family_id"`
	UserPic    string `gorm:"column:user_pic"`

	Login    string    `gorm:"column:login"`
	Password string    `gorm:"column:password"`
	Token    string    `gorm:"column:token"`
	LastAuth time.Time `gorm:"column:last_auth"`
	Metadata string    `gorm:"column:metadata"`
}

func (u *User) Create() error {

	res := db.Create(&u)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (u *User) Update() error {

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

// Family Read family
func (u *User) Family() ([]User, error) {
	var users []User

	res := db.Table("users").Where("family_id", u.FamilyId).Find(&users)
	return users, res.Error
}

// GetUser in db
func GetUser(chatId int64) *User {
	u := &User{TelegramId: chatId}
	u.read()
	return u
}

// Users array
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

// GetUsersList all users in db
func GetUserList() []int64 {
	u := &Users{}
	u.read()
	return u.list()
}

// Family
type Family struct {
	gorm.Model
	Owner  uint   `gorm:"column:owner"`
	Active string `gorm:"column:active"`
}

func (f *Family) Create() error {

	res := db.Create(&f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (f *Family) Update() error {

	res := db.Save(f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (f *Family) Read() error {

	res := db.Where(f).Find(&f)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// GetUser in db
func GetUserFamily(userId uint) *Family {
	f := &Family{Owner: userId}
	f.Read()
	return f
}
