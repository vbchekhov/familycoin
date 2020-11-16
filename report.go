package main

import (
	"database/sql"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"os"
	"strconv"
	"time"
)

/* Reports */

// reports handle
func reports(c *skeleton.Context) bool {

	kb := skeleton.NewInlineKeyboard(1, 5)
	kb.ChatID = c.ChatId()
	kb.Title = "Чавой тебе рассказать?"
	kb.Buttons.Add("💰 Казна", "rep_1")
	kb.Buttons.Add("📈 Последние приходы", "rep_2")
	kb.Buttons.Add("📉 Последние расходы", "rep_3")
	kb.Buttons.Add("📊 Выгрузить в excel", "export_excel")

	if c.RegexpResult[0] == "📊 Отчетность" {
		m := tgbotapi.NewMessage(c.ChatId(), kb.Title)
		m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
		c.BotAPI.Send(m)
	}

	if c.RegexpResult[0] == "back_to_reports" {
		m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
		m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
		c.BotAPI.Send(m)
	}

	return true
}

// current balance
func balance(c *skeleton.Context) bool {

	// back button menu reports
	kb := skeleton.NewInlineKeyboard(1, 1)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	var sum float64

	text := "🤴 В казне сейчас, милорд!\n"
	for _, b := range balances(c.ChatId()) {
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

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

/* Excel reports */

// ExcelData
type ExcelData []struct {
	Date      string `gorm:"column:date"`
	DebitCat  string `gorm:"column:debit_cat"`
	CreditCat string `gorm:"column:credit_cat"`
	DebitSum  int    `gorm:"column:debit_sum"`
	CreditSum int    `gorm:"column:credit_sum"`
	Comment   string `gorm:"column:comment"`
	UserName  string `gorm:"column:user_name"`
}

// read
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

// exportExcel
func exportExcel(c *skeleton.Context) bool {

	f := excelize.NewFile()
	f.NewSheet("Sheet1")

	// ----------------------------------------------------------------------------------
	// |   date   | debit-cat | credit-cat | debit-sum | credit-sum | username | comment |
	// |    A%d   |   B%d     |    C%d     |    D%d    |     E%d    |    F%d  |    G%d  |
	// ----------------------------------------------------------------------------------

	f.SetColWidth("Sheet1", "A", "F", 20)

	f.SetCellStr("Sheet1", "A1", "Дата")
	f.SetCellStr("Sheet1", "B1", "Категория 'Пришло'")
	f.SetCellStr("Sheet1", "C1", "Категория 'Ушло'")
	f.SetCellStr("Sheet1", "D1", "Сумма 'Пришло'")
	f.SetCellStr("Sheet1", "E1", "Сумма 'Ушло'")
	f.SetCellStr("Sheet1", "F1", "Записал")
	f.SetCellStr("Sheet1", "G1", "Комментарий")

	u := &User{TelegramId: c.ChatId()}
	u.read()

	var ed ExcelData
	ed.read(u)

	for i := 0; i < len(ed); i++ {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), ed[i].Date)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), ed[i].DebitCat)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), ed[i].CreditCat)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), ed[i].DebitSum)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), ed[i].CreditSum)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), ed[i].UserName)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), ed[i].Comment)
	}

	f.SaveAs("./reports.xlsx")
	c.BotAPI.Send(tgbotapi.NewDocumentUpload(c.ChatId(), "./reports.xlsx"))
	os.Remove("./reports.xlsx")

	return true
}

/* Debit reports */

// list debit reports
func debitsReports(c *skeleton.Context) bool {

	today := time.Now()

	// create list report
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Title = "За сколько тебе показать приходы?"
	kb.Buttons.Add("📈 Приходы за 7 дней", "week_debit")
	kb.Buttons.Add("📈 Приходы за месяц", "month_debit")
	kb.Buttons.Add("📈 Приходы за "+monthf(today.Month()), "this_month_debit")
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// week debits
func weekDebit(c *skeleton.Context) bool {

	debit := &Debit{}
	text := debit.ReportDetail("последние 7 дней", c.ChatId(), time.Now().Add(-time.Hour*24*7), time.Now())

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// moth debits
func monthDebit(c *skeleton.Context) bool {

	debit := &Debit{}
	text := debit.ReportGroup("последний месяц", c.ChatId(), time.Now().Add(-time.Hour*24*30), time.Now())

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// this mouth debits
func thisMonthDebit(c *skeleton.Context) bool {

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	debit := &Debit{}
	text := debit.ReportGroup(monthf(today.Month()), c.ChatId(), start, end)

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

/* Credit reports */

// list credit reports
func creditsReports(c *skeleton.Context) bool {

	today := time.Now()
	// create list report
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Title = "За сколько тебе показать приходы?"
	kb.Buttons.Add("📉 Расходы за 7 дней", "week_credit")
	kb.Buttons.Add("📉 Расходы за месяц", "month_credit")
	kb.Buttons.Add("📉 Расходы за "+monthf(today.Month()), "this_month_credit")
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// week credits
func weekCredit(c *skeleton.Context) bool {

	credits := &Credit{}
	text := credits.ReportDetail("последние 7 дней", c.ChatId(), time.Now().Add(-time.Hour*24*7), time.Now())

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// month credits
func monthCredit(c *skeleton.Context) bool {

	credits := &Credit{}
	text := credits.ReportGroup("последний месяц", c.ChatId(), time.Now().Add(-time.Hour*24*30), time.Now())

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// this mouth debits
func thisMonthCredit(c *skeleton.Context) bool {

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	credits := &Credit{}
	text := credits.ReportGroup(monthf(today.Month()), c.ChatId(), start, end)

	// back button
	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

/* Other reports and func */

// receipt
func receipt(c *skeleton.Context) bool {

	r := &Receipts{}
	operationId, _ := strconv.Atoi(c.RegexpResult[2])
	if c.RegexpResult[1] == "debits" {
		debit := &Debit{}
		r = Receipt(debit, operationId)
	}

	if c.RegexpResult[1] == "credits" {
		credit := &Credit{}
		r = Receipt(credit, operationId)

		credit.ID = uint(operationId)
		credit.read()

		if credit.Receipt != "" {
			UploadPhoto(c.BotAPI, c.ChatId(), credit.Receipt, r.messagef())
			return true
		}
	}

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, r.messagef())
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	return true
}

// monthf russian name
func monthf(mounth time.Month) string {

	months := map[time.Month]string{
		time.January:   "❄️ Январь",
		time.February:  "🌨 Февраль",
		time.March:     "💃 Март",
		time.April:     "🌸 Апрель",
		time.May:       "🕊 Май",
		time.June:      "🌞 Июнь",
		time.July:      "🍉 Июль",
		time.August:    "⛱ Август",
		time.September: "🍁 Сентябрь",
		time.October:   "🍂 Октябрь",
		time.November:  "🥶 Ноябрь",
		time.December:  "🎅 Декабрь",
	}

	return months[mounth]
}
