package main

import (
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

	if !userExist(c.ChatId()) {
		return true
	}

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

	if !userExist(c.ChatId()) {
		return true
	}

	// back button menu reports
	kb := skeleton.NewInlineKeyboard(1, 1)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, "🤴 В казне сейчас "+strconv.Itoa(balanceNow(c.ChatId()))+" рублей, милорд!")
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

/* Excel reports */

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

	if !userExist(c.ChatId()) {
		return true
	}

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	var text string = "***Приходы за последние 7 дней*** 📈\n\n"
	var sum int

	// get detail report
	ad := debitsDetail(c.ChatId(), time.Now().Add(-time.Hour*24*7), time.Now())
	for _, s := range ad {
		text += s.Created.Format("02.01") + " " + s.Name + ": " + strconv.Itoa(s.Sum) + " руб. _" + s.Comment + "_\n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	var text string = "***Приходы за последний месяц*** 📈\n\n"
	var sum int

	// get group report
	ad := debitsGroup(c.ChatId(), time.Now().Add(-time.Hour*24*30), time.Now())

	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	today := time.Now()
	var text string = "***Приходы за " + monthf(today.Month()) + "*** 📈\n\n"
	var sum int

	// get group report
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	ad := debitsGroup(c.ChatId(), start, end)

	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

	if !userExist(c.ChatId()) {
		return true
	}

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	var text string = "***Расходы за последние 7 дней*** 📉\n\n"
	var sum int

	// get detail report
	ac := creditsDetail(c.ChatId(), time.Now().Add(-time.Hour*24*7), time.Now())
	for _, s := range ac {
		text += s.Created.Format("02.01") + " " + s.Name + ": " + strconv.Itoa(s.Sum) + " руб. _" + s.Comment + "_\n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	var text string = "***Расходы за последний месяц*** 📉\n\n"
	var sum int

	// get detail report
	ad := creditsGroup(c.ChatId(), time.Now().Add(-time.Hour*24*30), time.Now())
	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

	if !userExist(c.ChatId()) {
		return true
	}

	// title
	today := time.Now()
	var text string = "***Расходы за " + monthf(today.Month()) + "*** 📉\n\n"
	var sum int

	// get group report
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	ad := creditsGroup(c.ChatId(), start, end)
	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	// total sum
	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

// send detail operation
func receipt(c *skeleton.Context) bool {

	// text detail
	operationDet := ""
	// type (debit / credit)
	operationType := c.RegexpResult[1]
	// id in table
	operationId, _ := strconv.Atoi(c.RegexpResult[2])

	if operationType == "debit" {
		// read opetation
		d := &Debit{}
		d.ID = uint(operationId)
		d.read()
		// read type name
		dt := DebitType{Id: d.DebitTypeId}
		dt.read()
		// create message
		operationDet = fmt.Sprintf("📝 Чек №%d\n"+
			"---\n"+
			"📍Cумма операции: %d\n"+
			"📍Комментарий: %s\n"+
			"📍Категория: %s",
			d.ID, d.Sum, d.Comment, dt.Name)
		// send message
		c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), operationDet))

		return true
	}

	if operationType == "credit" {
		// read opetation
		cr := &Credit{}
		cr.ID = uint(operationId)
		cr.read()
		// read type name
		ct := CreditType{Id: cr.CreditTypeId}
		ct.read()
		// create message
		operationDet = fmt.Sprintf("📝 Чек №%d\n"+
			"---\n"+
			"📍Cумма операции: %d\n"+
			"📍Комментарий: %s\n"+
			"📍Категория: %s",
			cr.ID, cr.Sum, cr.Comment, ct.Name)
		// send message
		if cr.Receipt == "" {
			c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(), operationDet))
			return true
		}
		// send message with photo
		if cr.Receipt != "" {
			UploadPhoto(c.BotAPI, c.ChatId(), cr.Receipt, operationDet)
			return true
		}
	}

	return false
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
