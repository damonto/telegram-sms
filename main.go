package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/damonto/telegram-sms/config"
	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
)

var Version string

func init() {
	dir, err := os.MkdirTemp("", "telegram-sms")
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.C.BotToken, "bot-token", "", "telegram bot token")
	flag.Int64Var(&config.C.AdminId, "admin-id", 0, "telegram admin id")
	flag.BoolVar(&config.C.IsEuicc, "euicc", false, "enable eUICC features")
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

	if !config.C.DontDownload && config.C.IsEuicc {
		lpac.Download(config.C.Dir, config.C.Version)
	}

	bot, err := gotgbot.NewBot(config.C.BotToken, nil)
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		panic(err)
	}

	_, err = modem.NewManager()
	if err != nil {
		slog.Error("failed to create modem manager", "error", err)
		panic(err)
	}

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
