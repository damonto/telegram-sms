package handler

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func Start() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		_, err := ctx.Bot().SendMessage(
			ctx,
			tu.Message(
				tu.ID(update.Message.From.ID),
				"Congratulations! You have successfully started the bot. 🎉",
			).WithReplyParameters(&telego.ReplyParameters{
				MessageID: update.Message.MessageID,
			}),
		)
		return err
	}
}
