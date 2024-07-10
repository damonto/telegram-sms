package handler

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

type SendHandler struct {
	handler
	phoneNumber string
}

const (
	StateSendAskPhoneNumber = "send_ask_phone_number"
	StateSendAskMessage     = "send_ask_message"
)

func HandleSendCommand(c telebot.Context) error {
	h := &SendHandler{}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.States(map[string]telebot.HandlerFunc{
		StateSendAskPhoneNumber: h.handlePhoneNumber,
		StateSendAskMessage:     h.handleMessage,
	})
	return h.handle(c)
}

func (h *SendHandler) handle(c telebot.Context) error {
	h.state.Next(StateSendAskPhoneNumber)
	return c.Send("Please send me the phone number you want to send the message to")
}

func (h *SendHandler) handlePhoneNumber(c telebot.Context) error {
	if len(c.Text()) < 3 {
		if err := c.Send("The phone number you provided is invalid. Please send me the correct phone number."); err != nil {
			return err
		}
	}

	h.state.Next(StateSendAskMessage)
	h.phoneNumber = c.Text()
	return c.Send("Please send me the message you want to send")
}

func (h *SendHandler) handleMessage(c telebot.Context) error {
	if err := h.modem.SendSMS(h.phoneNumber, c.Text()); err != nil {
		c.Send(fmt.Sprintf("Failed to send SMS to *%s*", h.phoneNumber), &telebot.SendOptions{
			ParseMode: telebot.ModeMarkdownV2,
		})
		h.stateManager.Done(c)
		return err
	}
	h.stateManager.Done(c)
	return c.Send(fmt.Sprintf("Your SMS has been sent to *%s*", h.phoneNumber), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}
