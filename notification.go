package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
)

// send notification all family
func sendNotificationByFamily(c *skeleton.Context, text string, operation string) {

	// get user id
	u := &User{TelegramId: c.ChatId()}
	u.read()

	// read family
	family, _ := u.family()

	// send notif
	for i := range family {

		// dont send myself
		if family[i].TelegramId == c.ChatId() {
			continue
		}

		m := tgbotapi.NewMessage(family[i].TelegramId, text+"\n _👾 Внес запись: "+u.FullName+"_")
		m.ParseMode = tgbotapi.ModeMarkdown

		// details button
		kb := skeleton.NewInlineKeyboard(1, 1)
		kb.Id = c.Update.Message.MessageID
		kb.ChatID = c.ChatId()
		kb.Buttons.Add("🔍 Детали", operation)
		m.ReplyMarkup = kb.Generate().InlineKeyboardMarkup()

		c.BotAPI.Send(m)
	}
}
