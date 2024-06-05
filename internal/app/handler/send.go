package handler

import (
	"fmt"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type SendHandler struct {
	data map[int64]string
}

const (
	SendSmsStatePhoneNumber = "phone_number"
	SendSmsStateMessage     = "message"
)

func NewSendHandler() ConversationHandler {
	return &SendHandler{
		data: make(map[int64]string, 1),
	}
}

func (h *SendHandler) Command() string {
	return "send"
}

func (h *SendHandler) Description() string {
	return "Send a message to a phone number"
}

func (h *SendHandler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "Please enter the phone number you want to send the message to", nil); err != nil {
		return err
	}
	return handlers.NextConversationState(SendSmsStatePhoneNumber)
}

func (h *SendHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		SendSmsStatePhoneNumber: h.handlePhoneNumber,
		SendSmsStateMessage:     h.handleMessage,
	}
}

func (h *SendHandler) handlePhoneNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "Please enter the message you want to send", nil); err != nil {
		return err
	}
	h.data[ctx.EffectiveChat.Id] = ctx.EffectiveMessage.Text
	return handlers.NextConversationState(SendSmsStateMessage)
}

func (h *SendHandler) handleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	message := ctx.EffectiveMessage.Text
	phoneNumber := h.data[ctx.EffectiveChat.Id]
	delete(h.data, ctx.EffectiveChat.Id)

	slog.Info("sending message", "phone_number", phoneNumber, "message", message)
	b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Your message has been sent to *%s*", phoneNumber), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	})
	return handlers.EndConversation()
}
