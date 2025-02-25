package router

import (
	"github.com/damonto/telegram-sms/internal/app/handler"
	th "github.com/mymmrac/telego/telegohandler"
)

type router struct {
	*th.BotHandler
}

func NewRouter(handler *th.BotHandler) *router {
	return &router{handler}
}

func (r *router) Register() {
	r.Handle(handler.Start(), th.CommandEqual("start"))
}
