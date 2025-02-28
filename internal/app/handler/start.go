package handler

import (
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type StartHandler struct {
	*Handler
}

func NewStartHandler() *StartHandler {
	h := new(StartHandler)
	return h
}

func (h *StartHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		_, err := h.Reply(ctx, update, util.EscapeText("Welcome to the SMS bot."), nil)
		return err
	}
}
