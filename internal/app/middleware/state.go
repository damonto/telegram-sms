package middleware

import (
	"github.com/damonto/telegram-sms/internal/pkg/state"
	"gopkg.in/telebot.v3"
)

func WrapState(state *state.StateManager) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			c.Set("state", state)
			return next(c)
		}
	}
}
