package app

import (
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/damonto/telegram-sms/internal/app/routes"
)

type App interface {
	Start()
}

type app struct {
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	updater    *ext.Updater
}

func NewApp(bot *gotgbot.Bot) App {
	dispatcher := ext.NewDispatcher(nil)
	updater := ext.NewUpdater(dispatcher, nil)

	return &app{
		bot:        bot,
		dispatcher: dispatcher,
		updater:    updater,
	}
}

func (a *app) registerCoreServices() {
	routes.NewRouter(a.bot, a.dispatcher).Register()
}

func (a *app) Start() {
	a.registerCoreServices()

	err := a.updater.StartPolling(a.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		slog.Error("failed to start polling", "error", err)
		return
	}

	a.updater.Idle()
}
