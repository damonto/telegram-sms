package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
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
	flag.StringVar(&config.C.APDUDriver, "apdu-driver", "at", "the APDU driver to use (at or qmi)")
	flag.StringVar(&config.C.Version, "version", "v2.0.2", "the version of lpac to download")
	flag.StringVar(&config.C.Dir, "dir", "/tmp/telegram-sms", "the directory to store lpac")
	flag.BoolVar(&config.C.DontDownload, "dont-download", false, "don't download lpac binary")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "enable verbose mode")
	flag.Parse()
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
	if !config.C.DontDownload {
		if err := lpac.Download(config.C.Dir, config.C.Version); err != nil {
			slog.Warn("failed to download lpac", "error", err)
		}
	}
	slog.Info("you are using", "version", Version)

	bot, err := telebot.NewBot(telebot.Settings{
		Token: config.C.BotToken,
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
	imei, _ := modem.GetImei()
	model, _ := modem.GetModel()
	device := fmt.Sprintf("%s (%s)", model, imei)
	slog.Info("new SMS received", "device", device, "operatorName", operatorName, "sender", sender, "text", text)

	template := `
%s
[*%s*] %s
%s
`
	if _, err := bot.Send(telebot.ChatID(config.C.AdminId), util.EscapeText(fmt.Sprintf(template, device, operatorName, sender, text)), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	}); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}
