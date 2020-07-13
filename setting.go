package main

import (
	"crypto/md5"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"time"
)

func referal(c *skeleton.Context) bool {

	u := &User{TelegramId: c.ChatId()}
	u.get()

	f := &Family{Owner: u.ID}
	f.get()

	if u.FamilyId != 0 && f.ID == 0 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Вы уже состоите в семье!"))
		return true
	}

	if u.FamilyId == 0 {

		f := &Family{Owner: u.ID}
		f.set()
		f.get()

		u.FamilyId = f.ID
		u.update()
	}

	h := md5.New()

	f = &Family{Owner: u.ID}
	f.get()
	f.Active = fmt.Sprintf("%x", h.Sum([]byte(time.Now().Format("05.999999999Z07:00"))))
	f.update()

	c.BotAPI.Send(tgbotapi.NewMessage(
		c.ChatId(),
		fmt.Sprintf("t.me/%s?start=%s", c.BotAPI.Self.UserName, f.Active)))

	return true
}
