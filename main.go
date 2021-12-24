package main

import (
	"familycoin/binance"
	"familycoin/mobile"
	"familycoin/models"
	"familycoin/web"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/vbchekhov/skeleton"
)

// configuration data
var conf *Config
var logger = logrus.New()

func main() {

	if _, err := os.Stat("img"); os.IsNotExist(err) {
		os.Mkdir("img", 0777)
	}

	if _, err := os.Stat("familycoin.log"); os.IsNotExist(err) && conf.Bot.Debug == false {
		file, _ := os.Create("familycoin.log")
		logger.SetOutput(file)
	}

	// db migrator
	models.Migrator()

	// start web server
	web.BotToken = conf.Bot.Token
	web.BotName = conf.Bot.Name
	web.SessionLife = time.Hour * 24
	web.SetLogger(logger)
	web.SetDebug(conf.Web.Debug)

	go web.StartWebServer(conf.Web.Portf(), conf.Web.CertSRT, conf.Web.CertKEY, conf.Web.IsTSL())

	// start mobile rest api
	mobile.SetLogger(logger)
	mobile.SetDebug(conf.Web.Debug)
	mobile.SetTokenPassword(conf.Mobile.TokenPwd)
	go mobile.NewRestApi(conf.Mobile.Portf(), conf.Mobile.CertSRT, conf.Mobile.CertKEY, conf.Mobile.IsTSL())

	// rate updater
	go models.RateUpdater()

	// logger
	skeleton.SetLogger(logger)
	skeleton.SetOwnerBot(conf.Bot.Owner)

	// create app
	app := skeleton.NewBot(conf.Bot.Token)

	// Read users in db
	users := models.GetUserList()

	// default message if rule not found
	skeleton.SetDefaultMessage("Ой! Не понял тебя, попробуй еще раз..")

	// - start message for register user and new user Family
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

	app.HandleFunc("📈 Курсы валют", currencyRates).Border(skeleton.Private).Methods(skeleton.Messages).AllowList().Load(users...)

	// show detail push notif if you state in Family
	app.HandleFunc(`receipt_(debits|credits)_(\d{0,})`, receipt).Border(skeleton.Private).Methods(skeleton.Callbacks).AllowList().Load(users...)

	// referralByFamily link for access Family
	app.HandleFunc("referralByFamily", referralByFamily).Border(skeleton.Private).Methods(skeleton.Callbacks)

	app.Debug()
	app.Run()

}

func init() {

	var err error
	conf, err = newConfig()
	if err != nil {
		logger.Errorf("Error Read config %v", err)
		return
	}

	binance.SetLogger(logger)

	err = binance.New(conf.Binance.Api, conf.Binance.Secret)
	if err != nil {
		logger.Errorf("Error open binance api %v", err)
		return
	}

	models.SetLogger(logger)

	err = models.NewDB(conf.DataBase.ConnToMariaDB())
	if err != nil {
		logger.Errorf("Error open database %v", err)
		return
	}

	// load regexp
	CompiledRegexp = regexp.MustCompile(GenerateRegexpBySynonyms())
}
