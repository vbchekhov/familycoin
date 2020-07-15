package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

// configuration data
var conf, _ = newConfig()

func main() {

	// if is first start...
	if conf.IsFirstRun {
		firstRun()
	}

	// create app
	app := skeleton.NewBot(conf.Bot.Token)

	// default message if rule not found
	skeleton.SetDefaultMessage("Ой! Не понял тебя, попробуй еще раз..")

	// - start message for register user and new user family
	app.HandleFunc("/start (.*)", startReferal).Border(skeleton.Private).Methods(skeleton.Commands)
	app.HandleFunc("/start", start).Border(skeleton.Private).Methods(skeleton.Commands)

	/* Debit handlers */

	// start debit command
	app.HandleFunc("💰 Прибыло", debit).Border(skeleton.Private).Methods(skeleton.Messages)
	// select debit type and create sum
	debitPipe := app.HandleFunc("deb_(.*)", debitWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	debitPipe = debitPipe.Func(debitSum).Timeout(time.Second * 60)

	// add new debit types
	debitTypePipe := app.HandleFunc(`add_debit_cat_(\d{0,})`, debitTypeAdd).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	debitTypePipe = debitTypePipe.Func(debitTypeSave).Timeout(time.Second * 60)

	/* Credit handlers */

	// start credit command
	app.HandleFunc("💸 Убыло", credit).Border(skeleton.Private).Methods(skeleton.Messages)
	// select credit type and create sum
	creditPipe := app.HandleFunc("cred_(.*)", creditWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	creditPipe = creditPipe.Func(creditSum).Timeout(time.Second * 60)

	// add new credit types
	creditTypePipe := app.HandleFunc(`add_credit_cat_(\d{0,})`, creditTypeAdd).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	creditTypePipe = creditTypePipe.Func(creditTypeSave).Timeout(time.Second * 60)

	/* Reports amd settings */

	// start report menu
	app.HandleFunc("📊 Отчетность и настройки", reports).Border(skeleton.Private).Methods(skeleton.Messages)
	// back to report menu
	app.HandleFunc("back_to_reports", reports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// balance (debit - credit)
	app.HandleFunc("rep_1", balance).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// debit reports for week and month
	app.HandleFunc("rep_2", debitsReports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("week_debit", weekDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("month_debit", monthDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// credit reports for week adn month
	app.HandleFunc("rep_3", creditsReports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("week_credit", weekCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("month_credit", monthCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)

	// show detail push notif if you state in family
	app.HandleFunc(`oper_(.*)_(\d{0,})`, detailOperation).Border(skeleton.Private).Methods(skeleton.Callbacks)

	// referral link for access family
	app.HandleFunc("referral", referral).Border(skeleton.Private).Methods(skeleton.Callbacks)

	// -- ОТЧЕТНОСТЬ

	app.Debug()
	app.Run()

}

func firstRun() {

	db.DropTableIfExists(User{}, Family{}, DebitType{}, Debit{}, CreditType{}, Credit{})
	db.CreateTable(User{}, Family{}, DebitType{}, Debit{}, CreditType{}, Credit{})

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

		dt.create()
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

		ct.create()
	}

	for i := range conf.Bot.Users {
		u := User{TelegramId: conf.Bot.Users[i]}
		u.create()
	}

	os.Mkdir("img", 0777)

	conf.IsFirstRun = false

	b, _ := yaml.Marshal(conf)
	ioutil.WriteFile("./app.yaml", b, os.ModePerm)

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

func startReferal(c *skeleton.Context) bool {

	f := &Family{Active: c.RegexpResult[1]}
	f.read()

	if f.Owner == 0 {
		c.BotAPI.Send(tgbotapi.NewMessage(
			c.ChatId(),
			"Оу! Ссылка больше не доступна 😒. Запросите заново у главы семейтсва."))
		return true
	}

	u := &User{TelegramId: c.ChatId()}
	u.read()

	u.FamilyId = f.ID

	if u.ID != 0 {
		u.update()
	} else {
		u.create()
	}

	f.Active = ""
	f.update()

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность и настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Добро пожаловать с семью! Привет от "+u.FullName)
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}
