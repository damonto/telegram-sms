package handler

import (
	"fmt"

	"github.com/damonto/telegram-sms/internal/pkg/conversation"
	"gopkg.in/telebot.v3"
)

type SendHandler struct {
	handler
	phoneNumber  string
	conversation conversation.Conversation
}

const (
	SendAskPhoneNumber = "send_ask_phone_number"
	SendAskMessage     = "send_ask_message"
)

func HandleSendCommand(c telebot.Context) error {
	h := &SendHandler{}
	h.setModem(c)
	h.conversation = conversation.New(c)
	h.conversation.Steps(map[string]telebot.HandlerFunc{
		SendAskPhoneNumber: h.handlePhoneNumber,
		SendAskMessage:     h.handleMessage,
	})
	return h.Handle(c)
}

func (h *SendHandler) Handle(c telebot.Context) error {
	h.conversation.Next(SendAskPhoneNumber)
	return c.Send("Please send me the phone number you want to send the message to")
}

func (h *SendHandler) handlePhoneNumber(c telebot.Context) error {
	if len(c.Text()) < 3 {
		if err := c.Send("The phone number you provided is invalid. Please send me the correct phone number."); err != nil {
			return err
		}
	}

	h.conversation.Next(SendAskMessage)
	h.phoneNumber = c.Text()
	return c.Send("Please send me the message you want to send")
}

func (h *SendHandler) handleMessage(c telebot.Context) error {
	if err := h.modem.SendSMS(h.phoneNumber, c.Text()); err != nil {
		c.Send(fmt.Sprintf("Failed to send SMS to *%s*", h.phoneNumber), &telebot.SendOptions{
			ParseMode: telebot.ModeMarkdownV2,
		})
		h.conversation.Done()
		return err
	}
	h.conversation.Done()
	return c.Send(fmt.Sprintf("Your SMS has been sent to *%s*", h.phoneNumber), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}
