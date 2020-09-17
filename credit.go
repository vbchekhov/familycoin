package main

import (
	"fmt"
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
	u := User{TelegramId: c.ChatId()}
	u.read()

	// write new credit note in map
	// with credit_type_id and user_id
	ct, _ := strconv.Atoi(c.RegexpResult[1])
	creditNote[c.ChatId()] = &Credit{
		CreditTypeId: ct,
		UserId:       u.ID,
		telegramId:   c.ChatId(),
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
	mc := regexp.MustCompile(`^(\d{0,})(?: руб| рублей|)(?:, (.*)|)$`)
	find := mc.FindStringSubmatch(text)

	// check regexp array
	if len(find) < 2 {
		m := tgbotapi.NewMessage(c.ChatId(), "Упс! Не нашел ни суммы, ни комметария. Еще раз.")
		m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
		c.BotAPI.Send(m)

		c.Pipeline().Repeat()

		return true
	}

	if find[1] == "" {
		m := tgbotapi.NewMessage(c.ChatId(), "Упс! Не нашел сумму 😕. Еще раз.")
		m.ReplyMarkup = skeleton.NewAbortPipelineKeyboard("⛔️ Отмена")
		c.BotAPI.Send(m)

		c.Pipeline().Repeat()

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
	limit := creditNote[c.ChatId()].limit
	// delete note in map
	delete(creditNote, c.ChatId())
	// stop pipeline
	c.Pipeline().Stop()

	limitText := ""
	if limit != nil {
		limitText = fmt.Sprintf("\n---\nПотрачено по ***%s***: %d\nЛимит %d", limit.Name, limit.Sum, limit.Limits)
	}
	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Ага, "+find[1]+" рублей. Записал 🖌📓"+limitText)

	// details button
	kb := skeleton.NewInlineKeyboard(1, 1)
	kb.Id = c.Update.Message.MessageID
	kb.ChatID = c.ChatId()
	kb.Buttons.Add("🔍 Детали", "receipt_credits_"+strconv.Itoa(int(operationId)))
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
	m.ParseMode = tgbotapi.ModeMarkdown
	c.BotAPI.Send(m)

	// send push notif
	go sendNotificationByFamily(c, "Убыло "+strconv.Itoa(sum)+" рублей. ",
		"receipt_credits_"+strconv.Itoa(int(operationId)))

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
