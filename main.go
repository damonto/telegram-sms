package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/damonto/telegram-sms/config"
	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/maltegrosse/go-modemmanager"
)

var Version string

func init() {
	dir, err := os.MkdirTemp("", "telegram-sms")
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.C.BotToken, "bot-token", "", "telegram bot token")
	flag.Int64Var(&config.C.AdminId, "admin-id", 0, "telegram admin id")
	flag.StringVar(&config.C.Version, "version", "v2.0.1", "the version of lpac to download")
	flag.StringVar(&config.C.Dir, "dir", dir, "the directory to store lpac")
	flag.BoolVar(&config.C.DontDownload, "dont-download", false, "don't download lpac binary")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "enable verbose mode")
	flag.Parse()
}

func main() {
	if config.C.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if os.Geteuid() != 0 {
		slog.Error("Please run as root")
		os.Exit(1)
	}

	slog.Info("You are using", "version", Version)

	if err := config.C.IsValid(); err != nil {
		slog.Error("config is invalid", "error", err)
		os.Exit(1)
	}

	if !config.C.DontDownload {
		lpac.Download(config.C.Dir, config.C.Version)
	}

	bot, err := gotgbot.NewBot(config.C.BotToken, nil)
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		panic(err)
	}

	manager, err := modem.NewManager()
	if err != nil {
		slog.Error("failed to create modem manager", "error", err)
		panic(err)
	}
	go manager.SubscribeMessaging(func(modem *modem.Modem, sms modemmanager.Sms) {
		subscribe(bot, modem, sms)
	})

	app := app.NewApp(bot)
	go func() {
		if err := app.Start(); err != nil {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

func subscribe(bot *gotgbot.Bot, modem *modem.Modem, sms modemmanager.Sms) {
	sender, _ := sms.GetNumber()
	operatorName, _ := modem.GetOperatorName()
	text, _ := sms.GetText()
	imei, _ := modem.GetImei()
	model, _ := modem.GetModel()
	device := fmt.Sprintf("%s (%s)", model, imei)
	slog.Info("SMS received", "device", device, "operatorName", operatorName, "sender", sender, "text", text)

	template := `
%s
[*%s*] %s
%s
`
	if _, err := bot.SendMessage(config.C.AdminId, util.EscapeText(fmt.Sprintf(template, device, operatorName, sender, text)), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	}); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}
