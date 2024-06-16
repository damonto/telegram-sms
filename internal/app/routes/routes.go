package routes

import (
	"github.com/damonto/telegram-sms/config"
	"github.com/damonto/telegram-sms/internal/app/handler"
	mmiddleware "github.com/damonto/telegram-sms/internal/app/middleware"
	"gopkg.in/telebot.v3/middleware"
)

func (r *Router) commands() map[string]string {
	return map[string]string{
		"/send":     "Send SMS to a phone number",
		"/modems":   "List all available modems",
		"/chip":     "Get eUICC chip information",
		"/download": "Download eSIM profile",
		"/profiles": "Manage eSIM profiles",
	}
}

func (r *Router) routes() {
	r.bot.Handle("/start", handler.HandleStartCommand)
	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId))

		g.Handle("/modems", handler.HandleModemsCommand)
	}

	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId))
		g.Use(mmiddleware.SelectModem(false))

		g.Handle("/send", handler.HandleSendCommand)
	}

	{
		g := r.bot.Group()
		g.Use(middleware.Whitelist(config.C.AdminId))
		g.Use(mmiddleware.SelectModem(true))

		g.Handle("/chip", handler.HandleChipCommand)
		g.Handle("/profiles", handler.HandleProfilesCommand)
		g.Handle("/download", handler.HandleDownloadCommand)
	}
}
