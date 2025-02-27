package router

import (
	"context"
	"log/slog"
	"slices"

	"github.com/damonto/telegram-sms/internal/app/handler"
	"github.com/damonto/telegram-sms/internal/app/middleware"
	"github.com/damonto/telegram-sms/internal/app/state"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type router struct {
	*th.BotHandler
	bot *telego.Bot
	mm  *modem.Manager
	sm  *state.StateManager
}

func NewRouter(bot *telego.Bot, handler *th.BotHandler, mm *modem.Manager) *router {
	return &router{bot: bot, BotHandler: handler, mm: mm, sm: state.NewStateManager(handler)}
}

func (r *router) Register() {
	r.sm.RegisterCallback(r.BotHandler)
	r.registerCommands()
	r.registerHandlers()
	r.sm.RegisterMessage(r.BotHandler)
}

func (r *router) registerCommands() {
	commands := []telego.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "modem", Description: "List all plugged in modems"},
		{Command: "slot", Description: "List all SIM slots on the modem"},
		{Command: "chip", Description: "Get the eUICC chip information"},
		{Command: "ussd", Description: "Send a USSD command to the carrier"},
		{Command: "send", Description: "Send an SMS to a phone number"},
		{Command: "profiles", Description: "List all profiles on the eUICC"},
		{Command: "download", Description: "Download a profile into the eUICC"},
	}

	if err := r.bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{
		Scope: &telego.BotCommandScopeAllPrivateChats{
			Type: telego.ScopeTypeAllPrivateChats,
		},
		Commands: commands,
	}); err != nil {
		slog.Error("Failed to set commands", "error", err)
	}
}

func (r *router) registerHandlers() {
	r.Handle(handler.NewStartHandler().Handle(), th.CommandEqual("start"))

	modemRequiredMiddleware := middleware.NewModemRequiredMiddleware(r.mm, r.BotHandler)

	admin := r.Group(th.Not(th.CommandEqual("start")))
	admin.Use(middleware.Admin())
	admin.Handle(handler.NewListModemHandler(r.mm).Handle(), th.CommandEqual("modem"))

	{
		standard := admin.Group(r.predicate([]string{"/send", "/slot", "/ussd", "/send"}))
		standard.Use(modemRequiredMiddleware.Middleware(false))
		standard.Handle(handler.NewSIMSlotHandler().Handle(), th.CommandEqual("slot"))
		standard.Handle(handler.NewUSSDHandler().Handle(), th.CommandEqual("ussd"))
		standard.Handle(handler.NewSendHandler().Handle(), th.CommandEqual("send"))
	}

	{
		euicc := admin.Group(r.predicate([]string{"/chip", "/profiles", "/download"}))
		euicc.Use(modemRequiredMiddleware.Middleware(true))
		euicc.Handle(handler.NewChipHandler().Handle(), th.CommandEqual("chip"))
		euicc.Handle(handler.NewProfileHandler().Handle(), th.CommandEqual("profiles"))
		euicc.Handle(handler.NewDownloadHandler().Handle(), th.CommandEqual("download"))
	}
}

func (r *router) predicate(filters []string) th.Predicate {
	return func(ctx context.Context, update telego.Update) bool {
		return slices.Contains(filters, update.Message.Text)
	}
}
