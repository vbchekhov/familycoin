package main

import (
	"familycoin/models"
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

	text := models.GetBalance(c.ChatId()).Balancef()

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, text)
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

// exportExcel
func exportExcel(c *skeleton.Context) bool {

	f := excelize.NewFile()
	f.NewSheet("Sheet1")

	// ----------------------------------------------------------------------------------------------
	// |   date   | debit-cat | credit-cat | debit-sum | credit-sum | currency | username | comment |
	// |    A%d   |   B%d     |    C%d     |    D%d    |     E%d    |    F%d   |    G%d   |    H%d  |
	// ---------------------------------------------------------------------------------------------

	f.SetColWidth("Sheet1", "A", "F", 20)

	f.SetCellStr("Sheet1", "A1", "Дата")
	f.SetCellStr("Sheet1", "B1", "Категория 'Пришло'")
	f.SetCellStr("Sheet1", "C1", "Категория 'Ушло'")
	f.SetCellStr("Sheet1", "D1", "Сумма 'Пришло'")
	f.SetCellStr("Sheet1", "E1", "Сумма 'Ушло'")
	f.SetCellStr("Sheet1", "F1", "Записал")
	f.SetCellStr("Sheet1", "G1", "Комментарий")

	u := models.GetUser(c.ChatId())

	var ed models.ExcelData
	ed.Read(u)
	ed.Exchange()

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

	debit := &models.Debit{}
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

	debit := &models.Debit{}
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

	debit := &models.Debit{}
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

	credits := &models.Credit{}
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

	credits := &models.Credit{}
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

	credits := &models.Credit{}
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

	r := &models.Receipts{}
	operationId, _ := strconv.Atoi(c.RegexpResult[2])
	if c.RegexpResult[1] == "debits" {
		debit := &models.Debit{}
		r = models.Receipt(debit, uint(operationId))
	}

	if c.RegexpResult[1] == "credits" {
		credit := &models.Credit{}
		r = models.Receipt(credit, uint(operationId))

		credit.ID = uint(operationId)
		credit.Read()

		if credit.Receipt != "" {
			UploadPhoto(c.BotAPI, c.ChatId(), credit.Receipt, r.Fullf())
			return true
		}
	}

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, r.Fullf())
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
