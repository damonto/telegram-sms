package router

import (
	"context"
	"log/slog"

	"github.com/damonto/telegram-sms/internal/app/handler"
	"github.com/damonto/telegram-sms/internal/app/middleware"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type router struct {
	*th.BotHandler
	bot   *telego.Bot
	mm    *modem.Manager
	modem *modem.Modem
}

func NewRouter(bot *telego.Bot, handler *th.BotHandler, mm *modem.Manager) *router {
	return &router{bot: bot, BotHandler: handler, mm: mm}
}

func (r *router) Register() {
	r.registerCommands()
	r.registerHandlers()
}

func (r *router) registerCommands() {
	commands := []telego.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "modems", Description: "List all modems"},
		{Command: "chip", Description: "Get the eUICC chip information"},
	}

	if err := r.bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{
		Scope: &telego.BotCommandScopeAllPrivateChats{
			Type: telego.ScopeTypeAllPrivateChats,
		},
		Commands: commands,
	}); err != nil {
		slog.Error("failed to set commands", "error", err)
	}
}

func (r *router) registerHandlers() {
	r.Handle(handler.Start(), th.CommandEqual("start"))

	modemRequiredMiddleware := middleware.NewModemRequiredMiddleware(r.mm, r.BotHandler)

	admin := r.Group()
	admin.Use(middleware.Admin())

	admin.Handle(handler.ListModem(r.mm), th.CommandEqual("modems"))

	modemGroup := admin.Group()
	modemGroup.Use(modemRequiredMiddleware.Middleware)

	modemGroup.Handle(handler.Chip(), th.CommandEqual("chip"))
}
