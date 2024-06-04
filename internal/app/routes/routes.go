package routes

import (
	"github.com/damonto/telegram-sms/internal/app/handler"
	"github.com/damonto/telegram-sms/internal/app/middleware"
)

func (r *Router) routes() {
	r.registerCommand(handler.NewStartHandler(), nil)
	{
		m := middleware.Use(middleware.RequiredAdmin)
		r.registerConverstaion(handler.NewSendHandler(), m)
	}
}
