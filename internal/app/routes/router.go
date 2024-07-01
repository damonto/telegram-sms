package routes

import (
	"github.com/damonto/telegram-sms/internal/pkg/state"
	"gopkg.in/telebot.v3"
)

type Router struct {
	bot   *telebot.Bot
	state *state.StateManager
}

func NewRouter(bot *telebot.Bot, state *state.StateManager) *Router {
	return &Router{
		bot:   bot,
		state: state,
	}
}

func (r *Router) Register() error {
	r.routes()

	commands := make([]telebot.Command, 0, len(r.commands()))
	for command, description := range r.commands() {
		commands = append(commands, telebot.Command{Text: command, Description: description})
	}
	return r.bot.SetCommands(commands)
}
