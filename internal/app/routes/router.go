package routes

import (
	"gopkg.in/telebot.v3"
)

type Router struct {
	bot *telebot.Bot
}

func NewRouter(bot *telebot.Bot) *Router {
	return &Router{
		bot: bot,
	}
}

func (r *Router) Setup() error {
	r.routes()

	commands := make([]telebot.Command, 0, len(r.commands()))
	for command, description := range r.commands() {
		commands = append(commands, telebot.Command{Text: command, Description: description})
	}
	return r.bot.SetCommands(commands)
}
