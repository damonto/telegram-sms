package handler

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

type USSDHandler struct {
	handler
}

const (
	StateUSSDExecuteCommand = "ussd_execute_command"
	StateUSSDRespondCommand = "ussd_respond_command"
)

func HandleUSSDCommand(c telebot.Context) error {
	h := &USSDHandler{}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.Stages(map[string]telebot.HandlerFunc{
		StateUSSDExecuteCommand: h.handleExecuteCommand,
		StateUSSDRespondCommand: h.handleRespondCommand,
	})
	return h.handle(c)
}

func (h *USSDHandler) handle(c telebot.Context) error {
	h.state.Next(StateUSSDExecuteCommand)
	return c.Send("Please send me the USSD command you want to execute")
}

func (h *USSDHandler) handleExecuteCommand(c telebot.Context) error {
	response, err := h.modem.RunUSSDCommand(c.Text())
	if err != nil {
		h.stateManager.Done(c)
		c.Send("Failed to execute USSD command, err: " + err.Error())
		return err
	}
	h.state.Next(StateUSSDRespondCommand)
	return c.Send(fmt.Sprintf("%s\n%s\nIf you want to respond to this USSD command, please send me the response.", c.Text(), response))
}

func (h *USSDHandler) handleRespondCommand(c telebot.Context) error {
	response, err := h.modem.RespondUSSDCommand(c.Text())
	if err != nil {
		h.stateManager.Done(c)
		c.Send("Failed to respond to USSD command, err: " + err.Error())
		return err
	}
	h.state.Next(StateUSSDRespondCommand)
	return c.Send(fmt.Sprintf("%s\n%s\nIf you want to respond to this USSD command, please send me the response.", c.Text(), response))
}
