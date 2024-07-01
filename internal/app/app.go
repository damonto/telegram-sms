package app

import (
	"log/slog"

	"github.com/damonto/telegram-sms/internal/app/routes"
	"github.com/damonto/telegram-sms/internal/pkg/state"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type App interface {
	Start()
}

type app struct {
	bot *telebot.Bot
}

func NewApp(bot *telebot.Bot) App {
	return &app{
		bot: bot,
	}
}

func (a *app) setup() error {
	a.bot.Use(middleware.Recover())
	a.bot.Use(middleware.AutoRespond())

	if err := routes.NewRouter(a.bot, state.NewState(a.bot)).Register(); err != nil {
		slog.Error("failed to setup router", "error", err)
		return err
	}
	return nil
}

func (a *app) Start() {
	if err := a.setup(); err != nil {
		panic(err)
	}
	a.bot.Start()
}
