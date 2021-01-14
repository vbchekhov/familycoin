package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
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

	m := tgbotapi.NewEditMessageText(
		c.ChatId(),
		c.Update.CallbackQuery.Message.MessageID,
		"Деньги пришли "+debitTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	// read user data
	u := GetUser(c.ChatId())

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
	note, err := TextToDebitCreditData(text)

	// check regexp array
	if err != nil && err.Error() == "Empty message\n" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел ни суммы, ни комметария. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	if err != nil && err.Error() == "Empty amount\n" {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Упс! Не нашел сумму 😕. Попробуй сначала."))

		c.Pipeline().Stop()
		return true
	}

	// if sum found, conv in int
	// and write sum
	debitNote[c.ChatId()].Sum = note.Sum
	debitNote[c.ChatId()].CurrencyTypeId = note.Currency.ID

	// check and write comment
	debitNote[c.ChatId()].Comment = note.Comment

	// create in base
	debitNote[c.ChatId()].create()

	// stop pipeline
	c.Pipeline().Stop()

	m := tgbotapi.NewMessage(
		c.ChatId(),
		fmt.Sprintf("Ага, пришло %d%s в казну.\n\n\n%s",
			note.Sum,
			note.Currency.SymbolCode,
			GetBalance(c.ChatId()).Balancef()),
	)
	m.ParseMode = tgbotapi.ModeMarkdown

	// details button
	m.ReplyMarkup = skeleton.NewInlineButton("🔍 Детали", debitNote[c.ChatId()].Receipts().OperationID())
	c.BotAPI.Send(m)

	go SendReceipts(c, debitNote[c.ChatId()])

	// delete note in map
	defer delete(debitNote, c.ChatId())

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
