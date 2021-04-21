package main

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"strconv"
)

/* Credit handlers */

// map credit notes
var creditNote = map[int64]*Credit{}

// map credit types
var creditTypes = map[string]string{}

// calculate the number of columns
func columns(count int) int {
	const max = 20

	if count%max == 0 {
		return count / max
	}

	return count/max%max + 1
}

// render credit type keyboard
func creditTypeKeyboard(chatId int64, messageId int) *tgbotapi.InlineKeyboardMarkup {

	// create map credit types
	// from database
	var ct = &CreditTypes{}
	creditTypes = ct.convmap()

	// create keyboard credit types
	kb := skeleton.NewInlineKeyboard(columns(len(creditTypes)+1), len(creditTypes)+1)
	kb.Id = messageId
	kb.ChatID = chatId
	for k, v := range creditTypes {
		kb.Buttons.Add(v, "cred_"+k)
	}
	// add credit type categories
	kb.Buttons.Add("➕ Добавить категорию", "add_credit_cat_"+strconv.Itoa(messageId+1))

	return kb.Generate().InlineKeyboardMarkup()
}

// start credit command
func credit(c *skeleton.Context) bool {

	m := tgbotapi.NewMessage(c.ChatId(), "Ну и куда ты протрЫнькал бабукати, кожанный ты мешок? 😡")
	m.ReplyMarkup = creditTypeKeyboard(c.ChatId(), c.Update.Message.MessageID)
	c.BotAPI.Send(m)

	return true
}

// create category in credit notes map
func creditWho(c *skeleton.Context) bool {

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID,
		"Ага, потратил на "+creditTypes[c.RegexpResult[1]]+"\nА сколько? 🤨")
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
	c.BotAPI.Send(m)

	// read user data
	u := GetUser(c.ChatId())

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

	// check text command
	text := c.Update.Message.Text
	if c.Update.Message.Photo != nil {
		text = c.Update.Message.Caption
	}

	// regexp message
	note, err := TextToDebitCreditData(text)

	// check regexp array
	if err != nil && err.Error() == "Empty message\n" {
		m := tgbotapi.NewMessage(c.ChatId(), "Упс! Не нашел ни суммы, ни комметария. Еще раз.")
		m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
		c.BotAPI.Send(m)

		c.Pipeline().Repeat()

		return true
	}

	if err != nil && err.Error() == "Empty amount\n" {
		m := tgbotapi.NewMessage(c.ChatId(), "Упс! Не нашел сумму 😕. Еще раз.")
		m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
		c.BotAPI.Send(m)

		c.Pipeline().Repeat()

		return true
	}

	// if sum found, conv in int
	// and write sum
	creditNote[c.ChatId()].Sum = note.Sum
	creditNote[c.ChatId()].CurrencyTypeId = note.Currency.ID

	// check and write comment
	creditNote[c.ChatId()].Comment = note.Comment

	// if photo found save in img/ dir,
	// and write not photo path
	if c.Update.Message.Photo != nil {
		photo := NewDownloadPhoto(c.BotAPI, *c.Update.Message.Photo, "img/", "")
		photo.Save()

		creditNote[c.ChatId()].Receipt = photo.Path()
	}

	// create in base
	creditNote[c.ChatId()].create()

	// stop pipeline
	c.Pipeline().Stop()

	m := tgbotapi.NewMessage(
		c.ChatId(),
		fmt.Sprintf("Ага, %s %s. Записал 🖌📓", note.Currency.FormatFunc(note.Sum), note.Currency.SymbolCode))
	m.ParseMode = tgbotapi.ModeMarkdown
	m.ReplyMarkup = skeleton.NewInlineButton("🔍 Детали", creditNote[c.ChatId()].Receipts().OperationID())
	c.BotAPI.Send(m)

	go SendReceipts(c, creditNote[c.ChatId()])

	defer delete(creditNote, c.ChatId())

	return true
}

// add new credit type
func creditTypeAdd(c *skeleton.Context) bool {

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

	// save credit type
	ct := &CreditType{Name: c.Update.Message.Text}
	ct.create()

	// create type in map
	creditTypes[strconv.Itoa(ct.Id)] = ct.Name

	// read message id
	messageId, _ := strconv.Atoi(c.Pipeline().Data()[0])

	// rebuild keyboard
	// send editing message
	c.BotAPI.Send(tgbotapi.NewEditMessageReplyMarkup(
		c.ChatId(),
		messageId,
		*creditTypeKeyboard(c.ChatId(), messageId)))

	// send notification if all ok
	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Новая категория "+c.Update.Message.Text+" добавлена! 👆"))

	c.Pipeline().Stop()

	return true
}

/* Credit handlers */
