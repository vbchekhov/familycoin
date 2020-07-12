package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"os"
	"skeleton"
	"time"
)

func main() {

	isFirstRun()

	// create app
	app := skeleton.NewBot("------")

	skeleton.SetDefaultMessage("Ой! Не понял тебя, попробуй еще раз..")

	// - приветствие +
	app.HandleFunc("/start", start).Border(skeleton.Private).Methods(skeleton.Commands)

	// -- ПРИХОДЫ --

	// - приход
	app.HandleFunc("💰 Прибыло", debit).Border(skeleton.Private).Methods(skeleton.Messages)
	//  - выбор вида прихода
	debitPipe := app.HandleFunc("deb_(.*)", debitWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	// - обработка суммы прихода
	debitPipe = debitPipe.Func(debitSum).Timeout(time.Second * 20)

	// -- ПРИХОДЫ --

	// -- РАСХОДЫ --

	app.HandleFunc("💸 Убыло", credit).Border(skeleton.Private).Methods(skeleton.Messages)
	//  - выбор вида расхода
	creditPipe := app.HandleFunc("cred_(.*)", creditWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	// - обработка суммы расхода
	creditPipe = creditPipe.Func(creditSum).Timeout(time.Second * 55)

	// -- РАСХОДЫ --

	// -- ОТЧЕТНОСТЬ

	app.HandleFunc("📊 Отчетность и настройки", reports).Border(skeleton.Private).Methods(skeleton.Messages)
	app.HandleFunc("rep_1", balance).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("rep_2", weekDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("rep_3", weekCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)

	// -- ОТЧЕТНОСТЬ

	app.Debug()
	app.Run()

}

func isFirstRun() bool {

	if _, err := os.Stat("img"); !os.IsNotExist(err) {
		return false
	}

	db.DropTableIfExists(User{}, DebitType{}, Debit{}, CreditType{}, Credit{})
	db.CreateTable(User{}, DebitType{}, Debit{}, CreditType{}, Credit{})

	var debitTypes = map[int]string{
		1: "👨‍🎨 От феодала (зп)",
		2: "🎅 По милости царя (проекты)",
		3: "🧏‍♂️За красивые глазки",
	}

	for i, s := range debitTypes {
		dt := &DebitType{
			Id:   i,
			Name: s,
		}

		dt.set()
	}

	var creditTypes = map[int]string{
		1:  "🥒 Полезная еда",
		2:  "🍟 Гадости (фастфуд)",
		3:  "🎬 Развекухи",
		4:  "🧖🏻‍♀️Красотища",
		5:  "🏠 Дом и все вот это",
		6:  "🚕 Покатухи",
		7:  "🎁 Подарочки",
		8:  "🛠🍀 Хобба",
		9:  "🧝🏼‍♂️Мой пиздюк",
		10: "👠👔 Шмотки",
	}

	for i, s := range creditTypes {
		ct := &CreditType{
			Id:   i,
			Name: s,
		}

		ct.set()
	}

	u := User{TelegramId: 0000000000}
	u.set()

	u = User{TelegramId: 0000000000}
	u.set()

	os.Mkdir("img", 0777)

	return true
}

func start(c *skeleton.Context) bool {

	if !userExist(c.ChatId()) {
		return true
	}

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность и настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Опять потратил денег, сука? 🙄")
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}
