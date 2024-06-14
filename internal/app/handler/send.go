package handler

import (
	"fmt"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type SendHandler struct {
	modemHandler
	data map[int64]string
}

const (
	SendSmsStatePhoneNumber = "sms_phone_number"
	SendSmsStateMessage     = "sms_message"
)

func NewSendHandler(dispatcher *ext.Dispatcher) ConversationHandler {
	h := &SendHandler{
		data: make(map[int64]string, 1),
	}
	h.dispathcer = dispatcher
	h.next = h.enter
	return h
}

func (h *SendHandler) Command() string {
	return "send"
}

func (h *SendHandler) Description() string {
	return "Send a message to a phone number"
}

func (h *SendHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		SendSmsStatePhoneNumber: h.handlePhoneNumber,
		SendSmsStateMessage:     h.handleMessage,
	}
}

func (h *SendHandler) enter(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveChat.Id, "Please enter the phone number you want to send the message to", nil); err != nil {
		return err
	}
	return handlers.NextConversationState(SendSmsStatePhoneNumber)
}

func (h *SendHandler) handlePhoneNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveChat.Id, "Please enter the message you want to send", nil); err != nil {
		return err
	}
	h.data[ctx.EffectiveChat.Id] = ctx.EffectiveMessage.Text
	return handlers.NextConversationState(SendSmsStateMessage)
}

func (h *SendHandler) handleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	message := ctx.EffectiveMessage.Text
	phoneNumber := h.data[ctx.EffectiveChat.Id]
	delete(h.data, ctx.EffectiveChat.Id)

	modem, err := h.modem(ctx)
	if err != nil {
		return err
	}
	if err := modem.SendSMS(phoneNumber, message); err != nil {
		b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Failed to send message to *%s*", phoneNumber), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeMarkdownV2,
		})
		return err
	}

	slog.Info("sending message", "phone_number", phoneNumber, "message", message)
	b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("Your message has been sent to *%s*", phoneNumber), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	})

	return handlers.EndConversation()
}
