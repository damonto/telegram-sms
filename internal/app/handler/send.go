package handler

import (
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

type SendHandler struct {
	phoneNumber string
}

const (
	ConverstationStatePhoneNumber = "phone_number"
	converstationStateMessage     = "message"
)

func NewSendHandler() ConversationHandler {
	return &SendHandler{}
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
	return handlers.NextConversationState(ConverstationStatePhoneNumber)
}

func (h *SendHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		ConverstationStatePhoneNumber: h.handlePhoneNumber,
		converstationStateMessage:     h.handleMessage,
	}
}

func (h *SendHandler) handlePhoneNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := ctx.EffectiveMessage.Reply(b, "Please enter the message you want to send", nil); err != nil {
		return err
	}
	h.phoneNumber = ctx.EffectiveMessage.Text
	return handlers.NextConversationState(converstationStateMessage)
}

func (h *SendHandler) handleMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	message := ctx.EffectiveMessage.Text
	slog.Info("message", "phone_number", h.phoneNumber, "message", message)
	// send the message
	return handlers.EndConversation()
}
