package main

import (
	"crypto/md5"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"strconv"
	"time"
)

/* Settings */

// settings
func settings(c *skeleton.Context) bool {

	kb := skeleton.NewInlineKeyboard(1, 2)
	kb.ChatID = c.ChatId()
	kb.Title = "⚙️ Настройки"
	kb.Buttons.Add("🧮 Добавить лимит расходов", "new_credit_limits")
	kb.Buttons.Add("👨‍👩‍👧 Добавить в семью", "referralByFamily")

	if c.RegexpResult[0] == "⚙️ Настройки" {
		m := tgbotapi.NewMessage(c.ChatId(), kb.Title)
		m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
		c.BotAPI.Send(m)
	}

	if c.RegexpResult[0] == "back_to_settings" {
		m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID, kb.Title)
		m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
		c.BotAPI.Send(m)
	}

	return true
}

// showCreditCategories
func showCreditCategories(c *skeleton.Context) bool {

	var ct = &CreditTypes{}
	creditTypes = ct.convmap()

	u := &User{TelegramId: c.ChatId()}
	u.read()

	// create keyboard credit types
	kb := skeleton.NewInlineKeyboard(columns(len(creditTypes)+1), len(creditTypes)+1)
	kb.Id = c.Update.CallbackQuery.Message.MessageID
	kb.ChatID = c.Update.CallbackQuery.Message.Chat.ID
	for k, v := range creditTypes {
		ctId, _ := strconv.Atoi(k)
		cl := &CreditLimit{
			CreditTypeId: ctId,
			UserId:       u.ID,
			FamilyId:     u.FamilyId,
		}
		cl.read()
		name := v
		if cl.ID != 0 {
			name += fmt.Sprintf(" (%d руб)", cl.Limit)
		}
		kb.Buttons.Add(name, "add_credit_limit_"+k)
	}
	kb.Buttons.Add("⬅️ Назад", "back_to_settings")

	m := tgbotapi.NewEditMessageText(c.ChatId(), c.Update.CallbackQuery.Message.MessageID,
		"✏️ Выбери категорию редактирования")
	m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()
	c.BotAPI.Send(m)

	return true
}

// editCreditLimit
func editCreditLimit(c *skeleton.Context) bool {

	id, _ := strconv.Atoi(c.RegexpResult[1])
	ct := &CreditType{Id: id}
	ct.read()

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(),
		"Введите лимит для "+ct.Name+"\n"+
			"Чтобы очистить лимит - напишите 0"))

	c.Pipeline().Save(c.RegexpResult[1])
	c.Pipeline().Save(ct.Name)
	c.Pipeline().Next()

	return true
}

// saveCreditLimit
func saveCreditLimit(c *skeleton.Context) bool {

	id, _ := strconv.Atoi(c.Pipeline().Data()[0])
	limit, _ := strconv.Atoi(c.Update.Message.Text)

	u := &User{TelegramId: c.Update.Message.Chat.ID}
	u.read()

	cl := &CreditLimit{
		CreditTypeId: id,
		UserId:       u.ID,
		FamilyId:     u.FamilyId,
	}
	cl.read()

	if limit == 0 {
		cl.delete()
		c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(),
			"Все, снова жизнь без лимитов!"))
		c.Pipeline().Stop()
		return true
	}

	if cl.ID == 0 {
		cl.Limit = limit
		cl.create()
	} else {
		cl.Limit = limit
		cl.update()
	}

	c.BotAPI.Send(tgbotapi.NewMessage(c.ChatId(),
		fmt.Sprintf("Добавлен лимит для %s в %s рублей",
			c.Pipeline().Data()[1],
			c.Update.Message.Text)))

	c.Pipeline().Stop()

	return true
}

// send referralByFamily link
func referralByFamily(c *skeleton.Context) bool {

	u := &User{TelegramId: c.ChatId()}
	u.read()

	f := &Family{Owner: u.ID}
	f.read()

	if u.FamilyId != 0 && f.ID == 0 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Вы уже состоите в семье!"))
		return true
	}

	if u.FamilyId == 0 {

		f := &Family{Owner: u.ID}
		f.create()
		f.read()

		u.FamilyId = f.ID
		u.update()
	}

	h := md5.New()

	f = &Family{Owner: u.ID}
	f.read()

	f.Active = fmt.Sprintf("%x", h.Sum([]byte(time.Now().Format("05.999999999Z07:00"))))
	f.update()

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Отправль эту ссылку своему члену семьи 👨‍👩‍👧 👇"))

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		fmt.Sprintf("t.me/%s?start=%s", c.BotAPI.Self.UserName, f.Active)))

	return true
}
