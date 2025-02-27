package handler

import (
	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Handler struct{}

type WithFunc = func(message *telego.SendMessageParams) error

func (h *Handler) LPA(ctx *th.Context) (*lpa.LPA, error) {
	return lpa.New(h.Modem(ctx))
}

func (h *Handler) Modem(ctx *th.Context) *modem.Modem {
	return ctx.Value("modem").(*modem.Modem)
}

func (h *Handler) reply(ctx *th.Context, to int64, text string, replyMessageId int, with WithFunc) (*telego.Message, error) {
	tm := tu.Message(
		tu.ID(to),
		text,
	).
		WithParseMode(telego.ModeMarkdownV2).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: replyMessageId,
		})
	if with != nil {
		if err := with(tm); err != nil {
			return nil, err
		}
	}
	message, err := ctx.Bot().SendMessage(ctx, tm)
	return message, err
}

func (h *Handler) Reply(ctx *th.Context, update telego.Update, text string, with WithFunc) (*telego.Message, error) {
	return h.reply(ctx, update.Message.Chat.ID, text, update.Message.MessageID, with)
}

func (h *Handler) ReplyMessage(ctx *th.Context, message telego.Message, text string, with WithFunc) (*telego.Message, error) {
	return h.reply(ctx, message.Chat.ID, text, message.MessageID, with)
}

func (h *Handler) ReplyCallbackQuery(ctx *th.Context, query telego.CallbackQuery, text string, with WithFunc) (*telego.Message, error) {
	return h.reply(ctx, query.From.ID, text, query.Message.GetMessageID(), with)
}
