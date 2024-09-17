package handler

import (
	"fmt"

	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type StartHandler struct{}

func HandleStartCommand(c telebot.Context) error {
	message := `
Hello, *%s %s*\!
Thanks for using this bot\.
Your UID is *%d*
	`
	return c.Send(fmt.Sprintf(message, util.EscapeText(c.Sender().FirstName), util.EscapeText(c.Sender().LastName), c.Sender().ID), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}
