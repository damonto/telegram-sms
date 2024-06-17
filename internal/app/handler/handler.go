package handler

import (
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"gopkg.in/telebot.v3"
)

type handler struct {
	modem *modem.Modem
}

func (h *handler) setModem(c telebot.Context) *modem.Modem {
	modem := c.Get("modem").(*modem.Modem)
	h.modem = modem
	return modem
}
