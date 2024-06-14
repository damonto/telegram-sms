package routes

import (
	"github.com/damonto/telegram-sms/internal/app/handler"
	"github.com/damonto/telegram-sms/internal/app/middleware"
)

func (r *Router) routes() {
	r.registerCommand(handler.NewStartHandler(), nil)
	{
		r.registerCommand(handler.NewModemHandler(), middleware.Use(middleware.RequiredAdmin))
		r.registerConverstaion(handler.NewSendHandler(r.dispatcher), middleware.Use(middleware.RequiredAdmin))
		r.registerConverstaion(handler.NewDownloadHandler(r.dispatcher), middleware.Use(middleware.RequiredAdmin))
		r.registerConverstaion(handler.NewProfileHandler(r.dispatcher), middleware.Use(middleware.RequiredAdmin))
	}
}
