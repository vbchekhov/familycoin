package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"strconv"
)

// -- ПРИХОДЫ --

var debitNote = map[int64]*Debit{}

var debitTypes = map[string]string{}

func debit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var dt = DebitTypes{}
	debitTypes = dt.convmap()

	kb := skeleton.NewInlineKeyboard(1, 3)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	for k, v := range debitTypes {
		kb.Buttons.Add(v, "deb_"+k)
	}

	m := tgbotapi.NewMessage(c.ChatId(),
		"Откуда бабукати? 🤑")
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

func debitWho(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	m := tgbotapi.NewMessage(c.ChatId(),
		"Деньги пришли "+debitTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	u := User{TelegramId: c.ChatId()}
	u.get()

	dt, _ := strconv.Atoi(c.RegexpResult[1])
	debitNote[c.ChatId()] = &Debit{
		DebitTypeId: dt,
		UserId:      u.ID,
	}

	c.Pipeline().Next()

	return true
}

func debitSum(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Ага, пришло "+c.Update.Message.Text+" рублей в казну.")
	m.ParseMode = tgbotapi.ModeMarkdown

	c.BotAPI.Send(m)

	sum, _ := strconv.Atoi(c.Update.Message.Text)
	debitNote[c.ChatId()].Sum = sum
	debitNote[c.ChatId()].set()

	delete(debitNote, c.ChatId())

	balance(c)

	c.Pipeline().Stop()

	return true
}

// -- ПРИХОДЫ --
