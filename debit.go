package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"regexp"
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

	var comment string

	text := c.Update.Message.Text

	mc := regexp.MustCompile(`^(\d{0,})(?: руб| рублей|)(?:, (.*)|)$`)
	find := mc.FindStringSubmatch(text)

	if len(find) < 2 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел ни суммы, ни комметария. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	if find[1] == "" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел сумму 😕. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	if len(find) == 3 {
		comment = find[2]
	}

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Ага, пришло "+c.Update.Message.Text+" рублей в казну.")
	m.ParseMode = tgbotapi.ModeMarkdown

	c.BotAPI.Send(m)

	sum, _ := strconv.Atoi(find[1])
	debitNote[c.ChatId()].Sum = sum
	debitNote[c.ChatId()].Comment = comment
	debitNote[c.ChatId()].set()

	delete(debitNote, c.ChatId())

	balance(c)

	c.Pipeline().Stop()

	return true
}

// -- ПРИХОДЫ --
