package routes

import (
	"github.com/damonto/telegram-sms/internal/app/handler"
	mmiddleware "github.com/damonto/telegram-sms/internal/app/middleware"
	"github.com/damonto/telegram-sms/internal/pkg/config"
	"gopkg.in/telebot.v3/middleware"
)

func (r *Router) commands() map[string]string {
	return map[string]string{
		"/send":     "Send SMS to a phone number",
		"/modems":   "List all available modems",
		"/chip":     "Get eUICC chip information",
		"/download": "Download eSIM profile",
		"/profiles": "Manage eSIM profiles",
		"/ussd":     "Execute USSD command",
		"/simslot":  "Manage SIM slots",
	}
}

func (r *Router) routes() {
	r.bot.Use(mmiddleware.WrapState(r.state))
	r.bot.Handle("/start", handler.HandleStartCommand)

	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId.ToInt64()...))

		g.Handle("/modems", handler.HandleModemsCommand)
	}

	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId.ToInt64()...))
		g.Use(mmiddleware.SelectModem(false))

		g.Handle("/send", handler.HandleSendCommand)
		g.Handle("/ussd", handler.HandleUSSDCommand)
		g.Handle("/simslot", handler.HandleSimSlotCommand)
	}

	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId.ToInt64()...))
		g.Use(mmiddleware.SelectModem(true))

		g.Handle("/chip", handler.HandleChipCommand)
		g.Handle("/profiles", handler.HandleProfilesCommand)
		g.Handle("/download", handler.HandleDownloadCommand)
	}
}
