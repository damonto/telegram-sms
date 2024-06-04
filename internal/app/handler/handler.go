package handler

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type Handler interface {
	Command() string
	Description() string
	Handle(bot *gotgbot.Bot, ctx *ext.Context) error
}

type ConversationHandler interface {
	Handler
	Conversations() map[string]handlers.Response
}
