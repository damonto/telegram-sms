package middleware

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type Middleware struct {
	next        handlers.Response
	middlewares []MiddlewareFunc
}

type MiddlewareFunc = func(bot *gotgbot.Bot, ctx *ext.Context) error

func Use(middlewares ...MiddlewareFunc) *Middleware {
	return &Middleware{
		middlewares: middlewares,
	}
}

func (m *Middleware) Next(next handlers.Response) *Middleware {
	m.next = next
	return m
}

func (m *Middleware) Handle(bot *gotgbot.Bot, ctx *ext.Context) error {
	for _, middleware := range m.middlewares {
		if err := middleware(bot, ctx); err != nil {
			bot.SendMessage(ctx.EffectiveChat.Id, err.Error(), nil)
			return err
		}
	}
	return m.next(bot, ctx)
}
