package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/damonto/libeuicc-go"
	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/maltegrosse/go-modemmanager"
	"gopkg.in/telebot.v3"
)

var Version string

func init() {
	if err := os.MkdirAll("/tmp/telegram-sms", 0755); err != nil {
		panic(err)
	}
	flag.StringVar(&config.C.BotToken, "bot-token", "", "telegram bot token")
	flag.Int64Var(&config.C.AdminId, "admin-id", 0, "telegram admin id")
	flag.StringVar(&config.C.Endpoint, "endpoint", "https://api.telegram.org", "telegram endpoint")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "enable verbose mode")
	flag.Parse()
}

func main1() {
	test := func() {
		l, err := lpa.New("/dev/cdc-wdm0", 1)
		if err != nil {
			slog.Error("failed to create LPA", "error", err)
			panic(err)
		}
		defer l.Close()
		fmt.Println(l.Download(context.Background(), &libeuicc.ActivationCode{
			SMDP:       "millicomelsalvador.validereachdpplus.com",
			MatchingId: "GENERICJOWMI-FAHTCU0-SKFMYPW6UIEFGRWC8GE933ITFAUVN63WMUVHFOWTS80",
		}, nil))
		// 304
		// fmt.Println(l.ProcessNotification(303, false))
		// fmt.Println(l.ProcessNotification(304, false))
	}
	test()
}

func main() {
	if config.C.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if os.Geteuid() != 0 {
		slog.Error("please run as root")
		os.Exit(1)
	}

	if err := config.C.IsValid(); err != nil {
		slog.Error("config is invalid", "error", err)
		os.Exit(1)
	}

	slog.Info("you are using", "version", Version)

	bot, err := telebot.NewBot(telebot.Settings{
		Token: config.C.BotToken,
		URL:   config.C.Endpoint,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		Poller: &telebot.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		panic(err)
	}

	mmgr, err := modem.NewManager()
	if err != nil {
		slog.Error("failed to create modem manager", "error", err)
		panic(err)
	}

	go func() {
		mmgr.SubscribeMessaging(func(modem *modem.Modem, sms modemmanager.Sms) {
			subscribe(bot, modem, sms)
		})
	}()

	go app.NewApp(bot).Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

func subscribe(bot *telebot.Bot, modem *modem.Modem, sms modemmanager.Sms) {
	sender, _ := sms.GetNumber()
	operatorName, _ := modem.GetOperatorName()
	text, _ := sms.GetText()
	slog.Info("new SMS received", "operatorName", operatorName, "sender", sender, "text", text)

	template := `
*\[%s\] \- %s*
%s
`
	if _, err := bot.Send(
		telebot.ChatID(config.C.AdminId),
		fmt.Sprintf(
			template,
			util.EscapeText(operatorName),
			util.EscapeText(sender),
			fmt.Sprintf("`%s`", util.EscapeText(text)),
		),
		&telebot.SendOptions{
			ParseMode:             telebot.ModeMarkdownV2,
			DisableWebPagePreview: true,
		}); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}
