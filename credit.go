package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"regexp"
	"strconv"
)

// -- РАСХОДЫ --

var creditNote = map[int64]*Credit{}

var creditTypes = map[string]string{}

func credit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var ct = CreditTypes{}
	creditTypes = ct.convmap()

	kb := skeleton.NewInlineKeyboard(1, 10)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()

	for k, v := range creditTypes {
		kb.Buttons.Add(v, "cred_"+k)
	}

	m := tgbotapi.NewMessage(c.ChatId(),
		"Ну и куда ты протрЫнькал бабукати, кожанный ты мешок? 😡")
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	return true
}

func creditWho(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	m := tgbotapi.NewMessage(c.ChatId(),
		"Ага, потратил на "+creditTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	u := User{TelegramId: c.ChatId()}
	u.get()

	ct, _ := strconv.Atoi(c.RegexpResult[1])
	creditNote[c.ChatId()] = &Credit{
		CreditTypeId: ct,
		UserId:       u.ID,
	}

	c.Pipeline().Next()

	return true
}

func creditSum(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	var comment, photoPath string
	var photo *Photo
	photoFound := false

	text := c.Update.Message.Text
	if c.Update.Message.Photo != nil {
		text = c.Update.Message.Caption
		photoFound = true
	}

	mc := regexp.MustCompile(`^(\d{0,})(?: руб| рублей|)(?:, (.*)|)$`)
	find := mc.FindStringSubmatch(text)

	if len(find) < 2 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел ни суммы, ни комметария. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	if len(find) == 3 {
		comment = find[2]
	}
	if photoFound {
		photo = NewDownloadPhoto(c.BotAPI, *c.Update.Message.Photo, "img/", "")
		photo.Save()
		photoPath = photo.Path()
	}

	if find[1] == "" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел сумму 😕. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		find[1]+" рублей?! ну ты и транжира!"))

	sum, _ := strconv.Atoi(find[1])
	creditNote[c.ChatId()].Sum = sum
	creditNote[c.ChatId()].Comment = comment
	creditNote[c.ChatId()].Receipt = photoPath
	creditNote[c.ChatId()].set()

	delete(creditNote, c.ChatId())

	c.Pipeline().Stop()

	return true
}

// -- РАСХОДЫ --
