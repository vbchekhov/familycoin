package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"regexp"
	"strconv"
)

/* Credit handlers */

// map credit notes
var creditNote = map[int64]*Credit{}

// map credit types
var creditTypes = map[string]string{}

// start credit command
func credit(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	// create map credit types
	// from database
	var ct = &CreditTypes{}
	creditTypes = ct.convmap()

	// create keyboard credit types
	kb := skeleton.NewInlineKeyboard(1, len(creditTypes)+1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	for k, v := range creditTypes {
		kb.Buttons.Add(v, "cred_"+k)
	}
	// add credit type categories
	kb.Buttons.Add("➕ Добавить категорию", "add_credit_cat_"+strconv.Itoa(c.Update.Message.MessageID+1))

	m := tgbotapi.NewMessage(c.ChatId(), "Ну и куда ты протрЫнькал бабукати, кожанный ты мешок? 😡")
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
	c.BotAPI.Send(m)

	return true
}

// create category in credit notes map
func creditWho(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	m := tgbotapi.NewMessage(c.ChatId(), "Ага, потратил на "+creditTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	// read user data
	u := User{TelegramId: c.ChatId()}
	u.read()

	// write new credit note in map
	// with credit_type_id and user_id
	ct, _ := strconv.Atoi(c.RegexpResult[1])
	creditNote[c.ChatId()] = &Credit{
		CreditTypeId: ct,
		UserId:       u.ID,
	}

	// create next pipeline command
	c.Pipeline().Next()

	return true
}

// save credit sum
func creditSum(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	// check text command
	text := c.Update.Message.Text
	if c.Update.Message.Photo != nil {
		text = c.Update.Message.Caption
	}

	// regexp message
	mc := regexp.MustCompile(`^(\d{0,})(?: руб| рублей|)(?:, (.*)|)$`)
	find := mc.FindStringSubmatch(text)

	// check regexp array
	if len(find) < 2 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел ни суммы, ни комметария. Еще раз."))
		return true
	}

	if find[1] == "" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел сумму 😕. Еще раз."))
		return true
	}

	// if sum found, conv in int
	// and write sum
	sum, _ := strconv.Atoi(find[1])
	creditNote[c.ChatId()].Sum = sum

	// check and write comment
	if len(find) >= 3 {
		creditNote[c.ChatId()].Comment = find[len(find)-1]
	}

	// if photo found save in img/ dir,
	// and write not photo path
	if c.Update.Message.Photo != nil {
		photo := NewDownloadPhoto(c.BotAPI, *c.Update.Message.Photo, "img/", "")
		photo.Save()

		creditNote[c.ChatId()].Receipt = photo.Path()
	}

	// create in base
	creditNote[c.ChatId()].create()
	// save id note
	operationId := creditNote[c.ChatId()].ID
	// delete note in map
	delete(creditNote, c.ChatId())
	// stop pipeline
	c.Pipeline().Stop()

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Ага, "+find[1]+" рублей. Записал 🖌📓")
	// details button
	kb := skeleton.NewInlineKeyboard(1, 1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("🔍 Детали", "oper_credit_"+strconv.Itoa(int(operationId)))
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
	c.BotAPI.Send(m)

	// send push notif
	go sendPushFamily(c, "Убыло "+strconv.Itoa(sum)+" рублей. ",
		"oper_credit_"+strconv.Itoa(int(operationId)))

	return true
}

// add new credit type
func creditTypeAdd(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Напиши название новой категории."))

	// save in pipeline message id
	c.Pipeline().Save(c.RegexpResult[1])
	c.Pipeline().Next()

	return true
}

// save new credit type
// and read inline keyboard
func creditTypeSave(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	// save credit type
	dt := &CreditType{Name: c.Update.Message.Text}
	dt.create()

	// create type in map
	creditTypes[strconv.Itoa(dt.Id)] = dt.Name

	// read message id
	messageId, _ := strconv.Atoi(c.Pipeline().Data()[0])

	// rebuild keyboard
	kb := skeleton.NewInlineKeyboard(1, len(creditTypes)+1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	for k, v := range creditTypes {
		kb.Buttons.Add(v, "deb_"+k)
	}
	kb.Buttons.Add("➕ Добавить категорию", "add_credit_cat_"+strconv.Itoa(c.Update.Message.MessageID))

	// send editing message
	c.BotAPI.Send(tgbotapi.NewEditMessageReplyMarkup(
		c.ChatId(),
		messageId,
		*kb.Generate().InlineKeyboardMarkup()))

	// send notification if all ok
	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Новая категория "+c.Update.Message.Text+" добавлена! 👆"))

	c.Pipeline().Stop()

	return true
}

/* Credit handlers */
