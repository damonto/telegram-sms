package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/damonto/telegram-sms/config"
	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
)

var Version string

func init() {
	cwd, _ := os.Getwd()
	flag.StringVar(&config.C.BotToken, "bot-token", "", "Telegram bot token")
	flag.Int64Var(&config.C.AdminId, "admin-id", 0, "Telegram admin id")
	flag.BoolVar(&config.C.IsEuicc, "euicc", false, "Enable eUICC features")
	flag.StringVar(&config.C.LpacVersion, "lpac-version", "2.0.1", "lpac version")
	flag.StringVar(&config.C.DataDir, "data-dir", filepath.Join(cwd, "data"), "Data directory")
	flag.BoolVar(&config.C.DontDownload, "dont-download", false, "Don't download lpac binary")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "Enable verbose mode")
	flag.Parse()
}

func main() {
	if config.C.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("You are using", "version", Version)

	if err := config.C.IsValid(); err != nil {
		slog.Error("config is invalid", "error", err)
		os.Exit(1)
	}

	if !config.C.DontDownload && config.C.IsEuicc {
		lpac.Download(config.C.DataDir, config.C.LpacVersion)
	}

	bot, err := gotgbot.NewBot(config.C.BotToken, nil)
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	// fmt.Println(modem.NewManager())

	app := app.NewApp(bot)
	go app.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
