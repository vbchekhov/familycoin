package main

import (
	"crypto/md5"
	"familycoin/models"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"time"
)

/* Settings */

// settings
func settings(c *skeleton.Context) bool {

	kb := skeleton.NewInlineKeyboard(1, 2)
	kb.ChatID = c.ChatId()
	kb.Title = "⚙️ Настройки"
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

// send referralByFamily link
func referralByFamily(c *skeleton.Context) bool {

	u := models.GetUser(c.ChatId())
	f := models.GetUserFamily(u.ID)

	if u.FamilyId != 0 && f.ID == 0 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Вы уже состоите в семье!"))
		return true
	}

	if u.FamilyId == 0 {

		f := models.GetUserFamily(u.ID)
		f.Create()
		f.Read()

		u.FamilyId = f.ID
		u.Update()
	}

	h := md5.New()

	f = &models.Family{Owner: u.ID}
	f.Read()

	f.Active = fmt.Sprintf("%x", h.Sum([]byte(time.Now().Format("05.999999999Z07:00"))))
	f.Update()

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		"Отправль эту ссылку своему члену семьи 👨‍👩‍👧 👇"))

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		fmt.Sprintf("t.me/%s?start=%s", c.BotAPI.Self.UserName, f.Active)))

	return true
}
