package handler

import (
	"github.com/damonto/telegram-sms/internal/pkg/lpa"
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

func (h *handler) GetLPA() (*lpa.LPA, error) {
	qmiDevice, err := h.modem.GetQMIDevice()
	if err != nil {
		return nil, err
	}
	slot, err := h.modem.GetPrimarySimSlot()
	if err != nil {
		return nil, err
	}
	return lpa.New(qmiDevice, int(slot))
}
