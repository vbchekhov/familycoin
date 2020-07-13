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
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("💰 Казна", "rep_1")
	kb.Buttons.Add("📈 Последние приходы", "rep_2")
	kb.Buttons.Add("📉 Последние расходы", "rep_3")
	kb.Buttons.Add("👨‍👩‍👧 Добавить в семью", "referal")

	m := tgbotapi.NewMessage(c.ChatId(),
		"Чавой тебе рассказать?")
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

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

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(), "В казне сейчас "+strconv.Itoa(bal)+" рублей, милорд!"))

	return true
}

func weekDebit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Приходы за последние 7 дней*** 📈\n\n"
	var sum int

	t1, t2 := time.Now().Add(-time.Hour*24*7), time.Now()

	ad := debitsForTime(t1, t2)
	for _, s := range ad {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб.\n"
		sum += s.Sum
	}

	text += "---\n _Итого:_ " + strconv.Itoa(sum) + " рублей."

	m := tgbotapi.NewMessage(c.ChatId(), text)
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	return true
}

func weekCredit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var text string = "***Расходы за последние 7 дней*** 📉\n\n"
	var sum int

	t1, t2 := time.Now().Add(-time.Hour*24*7), time.Now()

	ac := creditsForTime(t1, t2)
	for _, s := range ac {
		text += s.Name + ": " + strconv.Itoa(s.Sum) + " руб.\n"
		sum += s.Sum
	}

	text += "---\n _Итого:_ " + strconv.Itoa(sum) + " рублей."

	m := tgbotapi.NewMessage(c.ChatId(), text)
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	return true
}

// -- ОТЧЕТНОСТЬ
