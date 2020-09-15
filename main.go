package main

import (
	"github.com/Sirupsen/logrus"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/vbchekhov/skeleton"
	"os"
	"time"
)

// configuration data
var conf, _ = newConfig()
var logger = logrus.New()

func main() {

	if _, err := os.Stat("img"); os.IsNotExist(err) {
		os.Mkdir("img", 0777)
	}

	migrator()

	// logger
	file, _ := os.OpenFile("./familycoin.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	logger.SetOutput(file)
	skeleton.SetLogger(logger)
	skeleton.SetOwnerBot(conf.Bot.Owner)

	// create app
	app := skeleton.NewBot(conf.Bot.Token)

	// read users in db
	u := &Users{}
	u.read()
	users := u.list()

	// default message if rule not found
	skeleton.SetDefaultMessage("Ой! Не понял тебя, попробуй еще раз..")

	// - start message for register user and new user family
	app.HandleFunc("/start (.*)", startNewFamilyUser).Border(skeleton.Private).Methods(skeleton.Commands)
	app.HandleFunc("/start", start).Border(skeleton.Private).Methods(skeleton.Commands).AllowList().Load(users...)

	/* Debit handlers */

	// start debit command
	app.HandleFunc("💰 Прибыло", debit).Border(skeleton.Private).Methods(skeleton.Messages).AllowList().Load(users...)
	// select debit type and create sum
	debitPipe := app.HandleFunc("deb_(.*)", debitWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	debitPipe = debitPipe.Func(debitSum).Timeout(time.Second * 60)

	// add new debit types
	debitTypePipe := app.HandleFunc(`add_debit_cat_(\d{0,})`, debitTypeAdd).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	debitTypePipe = debitTypePipe.Func(debitTypeSave).Timeout(time.Second * 60)

	/* Credit handlers */

	// start credit command
	app.HandleFunc("💸 Убыло", credit).Border(skeleton.Private).Methods(skeleton.Messages).AllowList().Load(users...)
	// select credit type and create sum
	creditPipe := app.HandleFunc("cred_(.*)", creditWho).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	creditPipe = creditPipe.Func(creditSum).Timeout(time.Second * 60)

	// add new credit types
	creditTypePipe := app.HandleFunc(`add_credit_cat_(\d{0,})`, creditTypeAdd).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	creditTypePipe = creditTypePipe.Func(creditTypeSave).Timeout(time.Second * 60)

	/* Settings */
	app.HandleFunc("⚙️ Настройки", settings).Border(skeleton.Private).Methods(skeleton.Messages).AllowList().Load(users...)
	// back to setting menu
	app.HandleFunc("back_to_settings", settings).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc(`new_credit_limits`, showCreditCategories).Border(skeleton.Private).Methods(skeleton.Callbacks).AllowList()
	creditLimitPipe := app.HandleFunc(`add_credit_limit_(\d{0,})`, editCreditLimit).Border(skeleton.Private).Methods(skeleton.Callbacks).Append()
	creditLimitPipe = creditLimitPipe.Func(saveCreditLimit).Timeout(time.Second * 60)
	/* Reports amd settings */

	// start report menu
	app.HandleFunc("📊 Отчетность", reports).Border(skeleton.Private).Methods(skeleton.Messages).AllowList().Load(users...)
	// back to report menu
	app.HandleFunc("back_to_reports", reports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// balance (debit - credit)
	app.HandleFunc("rep_1", balance).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// debit reports for week and month
	app.HandleFunc("rep_2", debitsReports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("week_debit", weekDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("this_month_debit", thisMonthDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("month_debit", monthDebit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	// credit reports for week adn month
	app.HandleFunc("rep_3", creditsReports).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("week_credit", weekCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("this_month_credit", thisMonthCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("month_credit", monthCredit).Border(skeleton.Private).Methods(skeleton.Callbacks)
	app.HandleFunc("export_excel", exportExcel).Border(skeleton.Private).Methods(skeleton.Callbacks)

	// show detail push notif if you state in family
	app.HandleFunc(`receipt_(debits|credits)_(\d{0,})`, receipt).Border(skeleton.Private).Methods(skeleton.Callbacks).AllowList().Load(users...)

	// referralByFamily link for access family
	app.HandleFunc("referralByFamily", referralByFamily).Border(skeleton.Private).Methods(skeleton.Callbacks)

	app.Debug()
	app.Run()

}

func start(c *skeleton.Context) bool {

	kb := skeleton.NewReplyKeyboard(2)
	kb.Buttons.Add("💰 Прибыло")
	kb.Buttons.Add("💸 Убыло")
	kb.Buttons.Add("📊 Отчетность")
	kb.Buttons.Add("⚙️ Настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Опять потратил денег, сука? 🙄")
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}

func startNewFamilyUser(c *skeleton.Context) bool {

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
	kb.Buttons.Add("📊 Отчетность")
	kb.Buttons.Add("⚙️ Настройки")

	m := tgbotapi.NewMessage(
		c.ChatId(),
		"Добро пожаловать с семью! Привет от "+u.FullName)
	m.ReplyMarkup = kb.Generate().ReplyKeyboardMarkup()

	c.BotAPI.Send(m)

	return true

}
