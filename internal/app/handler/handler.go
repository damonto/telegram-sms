package handler

import (
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/state"
	"gopkg.in/telebot.v3"
)

type handler struct {
	modem        *modem.Modem
	stateManager *state.StateManager
	state        state.State
}

func (h *handler) init(c telebot.Context) {
	h.modem = c.Get("modem").(*modem.Modem)
	h.stateManager = c.Get("state").(*state.StateManager)
}
