package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"regexp"
	"strconv"
)

/* Debit handlers */

// map debit notes
var debitNote = map[int64]*Debit{}

// map debit types
var debitTypes = map[string]string{}

// debit type keyboard
func debitTypeKeyboard(chatId int64, messageId int) *tgbotapi.InlineKeyboardMarkup {

	// create map debit types
	// from database
	var dt = DebitTypes{}
	debitTypes = dt.convmap()

	// create keyboard debit types
	kb := skeleton.NewInlineKeyboard(columns(len(debitTypes)+1), len(debitTypes)+1)
	kb.Id = messageId
	kb.ChatID = chatId
	for k, v := range debitTypes {
		kb.Buttons.Add(v, "deb_"+k)
	}
	// add debit type categories
	kb.Buttons.Add("➕ Добавить категорию", "add_debit_cat_"+strconv.Itoa(messageId+1))

	return kb.Generate().InlineKeyboardMarkup()
}

// start credit command
func debit(c *skeleton.Context) bool {

	m := tgbotapi.NewMessage(c.ChatId(), "Откуда бабукати? 🤑")
	m.ReplyMarkup = debitTypeKeyboard(c.ChatId(), c.Update.Message.MessageID)

	c.BotAPI.Send(m)

	return true
}

// create category in credit notes map
func debitWho(c *skeleton.Context) bool {

	m := tgbotapi.NewMessage(c.ChatId(),
		"Деньги пришли "+debitTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	// read user data
	u := User{TelegramId: c.ChatId()}
	u.read()

	// write new debit note in map
	// with debit_type_id and user_id
	dt, _ := strconv.Atoi(c.RegexpResult[1])
	debitNote[c.ChatId()] = &Debit{
		DebitTypeId: dt,
		UserId:      u.ID,
	}

	// create next pipeline command
	c.Pipeline().Next()

	return true
}

// save credit sum
func debitSum(c *skeleton.Context) bool {

	// check text command
	text := c.Update.Message.Text

	// regexp message
	mc := regexp.MustCompile(`^(\d{0,})(?: руб| рублей|)(?:, (.*)|)$`)
	find := mc.FindStringSubmatch(text)

	// check regexp array
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

	// if sum found, conv in int
	// and write sum
	sum, _ := strconv.Atoi(find[1])
	debitNote[c.ChatId()].Sum = sum

	// check and write comment
	if len(find) >= 3 {
		debitNote[c.ChatId()].Comment = find[len(find)-1]
	}

	// create in base
	debitNote[c.ChatId()].create()
	// save id note
	operationId := debitNote[c.ChatId()].ID
	// delete note in map
	delete(debitNote, c.ChatId())
	// stop pipeline
	c.Pipeline().Stop()

	m := tgbotapi.NewMessage(c.ChatId(),
		"Ага, пришло "+c.Update.Message.Text+" рублей в казну.\n"+
			"Текущий баланс: "+strconv.Itoa(balanceNow(c.ChatId()))+" рублей.")
	m.ParseMode = tgbotapi.ModeMarkdown

	// details button
	kb := skeleton.NewInlineKeyboard(1, 1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("🔍 Детали", "oper_debit_"+strconv.Itoa(int(operationId)))
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

	c.BotAPI.Send(m)

	// send push notif
	go sendNotificationByFamily(c,
		"Поступило "+strconv.Itoa(sum)+" рублей. ",
		"oper_debit_"+strconv.Itoa(int(operationId)))

	return true
}

// add new credit type
func debitTypeAdd(c *skeleton.Context) bool {

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
func debitTypeSave(c *skeleton.Context) bool {

	// save debit type
	dt := &DebitType{Name: c.Update.Message.Text}
	dt.create()

	// create type in map
	debitTypes[strconv.Itoa(dt.Id)] = dt.Name

	// read message id
	messageId, _ := strconv.Atoi(c.Pipeline().Data()[0])

	// rebuild keyboard
	// send editing message
	c.BotAPI.Send(tgbotapi.NewEditMessageReplyMarkup(
		c.ChatId(),
		messageId,
		*debitTypeKeyboard(c.ChatId(), messageId)))

	// send notification if all ok
	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Новая категория "+c.Update.Message.Text+" добавлена! 👆"))

	c.Pipeline().Stop()

	return true
}

/* Debit handlers */
