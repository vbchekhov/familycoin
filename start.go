package main

import (
	"familycoin/models"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
)

// start
func start(c *skeleton.Context) bool {

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность")
	kb.Buttons.Add("📈 Курсы валют")
	kb.Buttons.Add("⚙️ Настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Опять потратил денег, сука? 🙄")
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	user := models.GetUser(c.ChatId())
	photos, _ := c.BotAPI.GetUserProfilePhotos(tgbotapi.NewUserProfilePhotos(int(c.ChatId())))
	photo := NewDownloadPhoto(c.BotAPI, photos.Photos[0], "img/", "")
	photo.Save()

	user.UserPic = photo.Path()
	user.Login = c.Update.Message.Chat.UserName
	user.FullName = c.Update.Message.Chat.FirstName + " " + c.Update.Message.Chat.LastName
	user.Update()

	c.BotAPI.Send(m)

	return true

}

// hello
func hello(c *skeleton.Context) bool {

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность")
	kb.Buttons.Add("📈 Курсы валют")
	kb.Buttons.Add("⚙️ Настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Ну привет")
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}

// startNewFamilyUser
func startNewFamilyUser(c *skeleton.Context) bool {

	f := &models.Family{Active: c.RegexpResult[1]}
	f.Read()

	if f.Owner == 0 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Оу! Ссылка больше не доступна 😒. Запросите заново у главы семейтсва."))
		return true
	}

	u := models.GetUser(c.ChatId())

	u.FamilyId = f.ID

	if u.ID != 0 {
		u.Update()
	} else {
		u.Create()
	}

	f.Active = ""
	f.Update()

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность")
	kb.Buttons.Add("⚙️ Настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Добро пожаловать с семью! Привет от "+u.FullName)
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}
