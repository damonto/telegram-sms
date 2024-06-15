package routes

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/damonto/telegram-sms/internal/app/handler"
	"github.com/damonto/telegram-sms/internal/app/middleware"
)

type Router struct {
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	commands   map[string]string
}

func NewRouter(bot *gotgbot.Bot, dispatcher *ext.Dispatcher) *Router {
	return &Router{
		bot:        bot,
		dispatcher: dispatcher,
		commands:   make(map[string]string),
	}
}

func (r *Router) addCommand(handler handler.CommandHandler) {
	r.commands[handler.Command()] = handler.Description()
}

func (r *Router) Register() {
	r.routes()

	var commands []gotgbot.BotCommand
	for command, description := range r.commands {
		commands = append(commands, gotgbot.BotCommand{Command: command, Description: description})
	}
	r.bot.SetMyCommands(commands, nil)
}

func (r *Router) registerConverstaion(handler handler.ConversationHandler, middleware *middleware.Middleware) {
	r.addCommand(handler)
	converstations := make(map[string][]ext.Handler, len(handler.Conversations()))
	for state, r := range handler.Conversations() {
		converstations[state] = []ext.Handler{
			handlers.NewMessage(func(msg *gotgbot.Message) bool {
				return message.Text(msg) && !message.Command(msg)
			}, r),
		}
	}
	r.dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand(handler.Command(), r.handler(handler, middleware))},
		converstations,
		&handlers.ConversationOpts{
			Exits: []ext.Handler{handlers.NewCommand("cancel", func(b *gotgbot.Bot, ctx *ext.Context) error {
				ctx.EffectiveMessage.Reply(b, "Cancelled", nil)
				return handlers.EndConversation()
			})},
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
			AllowReEntry: true,
		},
	))
}

func (r *Router) registerCommand(handler handler.CommandHandler, middleware *middleware.Middleware) {
	r.addCommand(handler)
	r.dispatcher.AddHandler(handlers.NewCommand(handler.Command(), r.handler(handler, middleware)))
}

func (r *Router) handler(handler handler.CommandHandler, middleware *middleware.Middleware) handlers.Response {
	if middleware != nil {
		return middleware.Next(handler.Handle).Handle
	}
	return handler.Handle
}
