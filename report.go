package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"strconv"
	"time"
)

// -- ОТЧЕТНОСТЬ

func reports(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	kb := skeleton.NewInlineKeyboard(1, 10)
	// kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Title = "Чавой тебе рассказать?"
	kb.Buttons.Add("💰 Казна", "rep_1")
	kb.Buttons.Add("📈 Последние приходы", "rep_2")
	kb.Buttons.Add("📉 Последние расходы", "rep_3")
	kb.Buttons.Add("👨‍👩‍👧 Добавить в семью", "referal")

	if c.RegexpResult[0] == "📊 Отчетность и настройки" {
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

func balance(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

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

	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, "🤴 В казне сейчас "+strconv.Itoa(bal)+" рублей, милорд!")
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

func debitsReports(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Title = "За сколько тебе показать приходы?"
	kb.Buttons.Add("📈 Приходы за 7 дней", "week_debit")
	kb.Buttons.Add("📈 Приходы за месяц", "month_debit")
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

func weekDebit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Приходы за последние 7 дней*** 📈\n\n"
	var sum int

	ad := debitForLastWeek()
	for _, s := range ad {
		text += s.Created.Format("02.01") + " " + s.Name + ": " + strconv.Itoa(s.Sum) + " руб. _" + s.Comment + "_\n"
		sum += s.Sum
	}

	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

func monthDebit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Приходы за последний месяц*** 📈\n\n"
	var sum int

	ad := debitsForTime(time.Now().Add(-time.Hour*24*30), time.Now())

	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

func creditsReports(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Title = "За сколько тебе показать приходы?"
	kb.Buttons.Add("📉 Расходы за 7 дней", "week_credit")
	kb.Buttons.Add("📉 Расходы за месяц", "month_credit")
	kb.Buttons.Add("⬅️ Назад", "back_to_reports")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

func weekCredit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Расходы за последние 7 дней*** 📉\n\n"
	var sum int

	ac := creditForLastWeek()
	for _, s := range ac {
		text += s.Created.Format("02.01") + " " + s.Name + ": " + strconv.Itoa(s.Sum) + " руб. _" + s.Comment + "_\n"
		sum += s.Sum
	}

	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

func monthCredit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Расходы за последний месяц*** 📉\n\n"
	var sum int

	ad := creditsForTime(time.Now().Add(-time.Hour*24*30), time.Now())

	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб. \n"
		sum += s.Sum
	}

	text += "---\n_Итого:_ " + strconv.Itoa(sum) + " рублей."

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

// -- ОТЧЕТНОСТЬ
