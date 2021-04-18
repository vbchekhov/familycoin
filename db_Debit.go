package main

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

/* Type working with Debits */

// Debit note
type Debit struct {
	gorm.Model
	DebitTypeId    int    `gorm:"column:debit_type_id" gorm:"association_foreignkey: id"`
	UserId         uint   `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum            int    `gorm:"column:sum"`
	Comment        string `gorm:"column:comment"`
	CurrencyTypeId uint   `gorm:"column:currency_type_id" gorm:"association_foreignkey: id"`
}

// ++ DebitCredit interface methods

func (d *Debit) Title() string {
	return "💰 Приходы"
}

func (d *Debit) ReceiptFile() string {
	return ""
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

func (d *Debit) ReportTitle(title string) string {
	return "***Приходы за " + title + "*** 📈\n\n"
}
func (d *Debit) ReportDetail(title string, chatId int64, start, end time.Time) string {

	// title
	text := d.ReportTitle(title)
	text += Detail(d, chatId, start, end).Detailsf()

	return text
}
func (d *Debit) ReportGroup(title string, chatId int64, start, end time.Time) string {

	text := d.ReportTitle(title)
	text += Group(d, chatId, start, end).Groupsf()

	return text
}

func (d *Debit) Receipts() *Receipts {
	return Receipt(d, d.ID)
}

// -- DebitCredit interface methods

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

// DebitType note
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

// DebitTypes array
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
