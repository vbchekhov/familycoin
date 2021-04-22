package main

import (
	"familycoin/models"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
)

// SendReceipts by Family
func SendReceipts(c *skeleton.Context, dt models.DebitCredit) {

	// get user id
	user := models.GetUser(c.ChatId())

	// Read Family
	family, _ := user.Family()

	// send notif
	for i := range family {

		// dont send myself
		if family[i].TelegramId == c.ChatId() {
			continue
		}

		m := tgbotapi.NewMessage(family[i].TelegramId, dt.Receipts().Shortf()+"\n _👾 Внес запись: "+user.FullName+"_")
		m.ParseMode = tgbotapi.ModeMarkdown
		m.ReplyMarkup = skeleton.NewInlineButton("🔍 Детали", dt.Receipts().OperationID())

		c.BotAPI.Send(m)

	}
}
