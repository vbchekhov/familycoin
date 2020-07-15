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

	kb := skeleton.NewInlineKeyboard(1, len(creditTypes)+1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()

	for k, v := range creditTypes {
		kb.Buttons.Add(v, "cred_"+k)
	}
	kb.Buttons.Add("➕ Добавить категорию", "add_credit_cat_"+strconv.Itoa(c.Update.Message.MessageID+1))

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
			"Упс! Не нашел ни суммы, ни комметария. Еще раз."))

		// c.Pipeline().Stop()
		return true
	}

	if find[1] == "" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел сумму 😕. Еще раз."))

		// c.Pipeline().Stop()
		return true
	}

	if len(find) >= 3 {
		comment = find[len(find)-1]
	}
	if photoFound {
		photo = NewDownloadPhoto(c.BotAPI, *c.Update.Message.Photo, "img/", "")
		photo.Save()
		photoPath = photo.Path()
	}
	sum, _ := strconv.Atoi(find[1])
	creditNote[c.ChatId()].Sum = sum
	creditNote[c.ChatId()].Comment = comment
	creditNote[c.ChatId()].Receipt = photoPath
	creditNote[c.ChatId()].set()

	operId := int(creditNote[c.ChatId()].ID)

	delete(creditNote, c.ChatId())

	c.Pipeline().Stop()

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		find[1]+" рублей?! ну ты и транжира!"))

	go sendPushFamily(c, "Убыло "+strconv.Itoa(sum)+" рублей. ", "oper_credit_"+strconv.Itoa(operId))

	return true
}

func creditTypeAdd(c *skeleton.Context) bool {

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Напиши название новой категории."))

	c.Pipeline().Save(c.RegexpResult[1])
	c.Pipeline().Next()

	return true
}

func creditTypeSave(c *skeleton.Context) bool {

	dt := CreditType{Name: c.Update.Message.Text}
	dt.set()

	creditTypes[strconv.Itoa(dt.Id)] = dt.Name

	mId, _ := strconv.Atoi(c.Pipeline().Data()[0])

	kb := skeleton.NewInlineKeyboard(1, len(debitTypes)+1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	for k, v := range creditTypes {
		kb.Buttons.Add(v, "deb_"+k)
	}
	kb.Buttons.Add("➕ Добавить категорию", "add_credit_cat_"+strconv.Itoa(c.Update.Message.MessageID))

	c.BotAPI.Send(tgbotapi.NewEditMessageReplyMarkup(
		c.ChatId(),
		mId,
		*kb.Generate().InlineKeyboardMarkup()))

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Новая категория "+c.Update.Message.Text+" добавлена! 👆"))

	c.Pipeline().Stop()

	return true
}

// -- РАСХОДЫ --
