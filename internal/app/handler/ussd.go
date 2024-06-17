package handler

import (
	"fmt"

	"github.com/damonto/telegram-sms/internal/pkg/conversation"
	"gopkg.in/telebot.v3"
)

type USSDHandler struct {
	handler
	conversation conversation.Conversation
}

const (
	USSDExecuteCommand = "ussd_execute_command"
	USSDRespondCommand = "ussd_respond_command"
)

func HandleUSSDCommand(c telebot.Context) error {
	h := &USSDHandler{}
	h.setModem(c)
	h.conversation = conversation.New(c)
	h.conversation.Flow(map[string]telebot.HandlerFunc{
		USSDExecuteCommand: h.handleExecuteCommand,
		USSDRespondCommand: h.handleRespondCommand,
	})
	return h.handle(c)
}

func (h *USSDHandler) handle(c telebot.Context) error {
	h.conversation.Next(USSDExecuteCommand)
	return c.Send("Please send me the USSD command you want to execute")
}

func (h *USSDHandler) handleExecuteCommand(c telebot.Context) error {
	response, err := h.modem.RunUSSDCommand(c.Text())
	if err != nil {
		c.Send("Failed to execute USSD command, err: " + err.Error())
		return err
	}
	h.conversation.Next(USSDRespondCommand)
	return c.Send(fmt.Sprintf("%s\n%s\nIf you want to respond to this USSD command, please send me the response.", c.Text(), response))
}

func (h *USSDHandler) handleRespondCommand(c telebot.Context) error {
	response, err := h.modem.RespondUSSDCommand(c.Text())
	if err != nil {
		c.Send("Failed to respond to USSD command, err: " + err.Error())
		return err
	}
	h.conversation.Next(USSDRespondCommand)
	return c.Send(fmt.Sprintf("%s\n%s\nIf you want to respond to this USSD command, please send me the response.", c.Text(), response))
}
