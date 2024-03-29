package models

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

/* Type working with Credits */

// Credit note
type Credit struct {
	gorm.Model
	CreditTypeId   int     `gorm:"column:credit_type_id" gorm:"association_foreignkey: id"`
	UserId         uint    `gorm:"column:user_id" gorm:"association_foreignkey: id"`
	Sum            float64 `gorm:"column:sum"`
	Comment        string  `gorm:"column:comment"`
	CurrencyTypeId uint    `gorm:"column:currency_type_id" gorm:"association_foreignkey: id"`
	Receipt        string  `gorm:"column:receipt"`
}

// ++ DebitCredit interface methods

func (c *Credit) Title() string {
	return "💸 Расходы"
}

func (c *Credit) ReceiptFile() string {
	return "d.Receipt as receipt,"
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

func (c *Credit) ReportTitle(title string) string {
	return "***Расходы за " + title + "*** 📉\n\n"
}
func (c *Credit) ReportDetail(title string, chatId int64, start, end time.Time) string {

	// title
	text := c.ReportTitle(title)
	text += Detail(c, chatId, start, end).Detailsf()

	return text
}
func (c *Credit) ReportGroup(title string, chatId int64, start, end time.Time) string {

	// title
	text := c.ReportTitle(title)
	text += Group(c, chatId, start, end).Groupsf()

	return text
}

func (c *Credit) Receipts() *Receipts {
	return Receipt(c, c.ID)
}

// -- DebitCredit interface methods

func (c *Credit) Create() error {

	res := db.Create(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *Credit) Read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

// CreditType note
type CreditType struct {
	Id   int    `gorm:"column:id" gorm:"primary_key" gorm:"AUTO_INCREMENT"`
	Name string `gorm:"column:name"`
}

func (c *CreditType) Create() error {

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

// CreditTypes array
type CreditTypes []CreditType

func (c *CreditTypes) read() error {

	res := db.Where(c).Find(&c)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
func (c *CreditTypes) Map() (m map[string]string) {
	m = make(map[string]string)

	c.read()

	for _, creditType := range *c {
		m[strconv.Itoa(creditType.Id)] = creditType.Name
	}

	return
}

func GetCreditTypes() CreditTypes {
	ct := CreditTypes{}
	ct.read()
	return ct
}
