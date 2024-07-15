package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type SimSlotHandler struct {
	handler
}

func HandleSimSlotCommand(c telebot.Context) error {
	h := &SimSlotHandler{}
	h.init(c)
	return h.handle(c)
}

func (h *SimSlotHandler) handle(c telebot.Context) error {
	simSlots, err := h.modem.GetSimSlots()
	if err != nil {
		return err
	}

	template := `
%s SIM %d - *%s*
%s
	`
	var text string
	selector := &telebot.ReplyMarkup{}
	buttons := make([]telebot.Btn, 0, len(simSlots))
	for slotId, simSlot := range simSlots {
		active, err := h.modem.GetSimActiveStatus(simSlot.GetObjectPath())
		if err != nil {
			return err
		}
		identifier, _ := simSlot.GetSimIdentifier()
		operatorName, _ := simSlot.GetOperatorName()
		if active {
			text += fmt.Sprintf(template, "🟢", slotId+1, operatorName, identifier)
		} else {
			text += fmt.Sprintf(template, "🔴", slotId+1, operatorName, identifier)
		}
		btn := selector.Data(fmt.Sprintf("SIM %d (%s)", slotId+1, identifier), fmt.Sprint(time.Now().UnixNano()), fmt.Sprint(slotId+1))
		c.Bot().Handle(&btn, func(c telebot.Context) error {
			return h.handleActiveSimSlot(c)
		})
		buttons = append(buttons, btn)
	}
	selector.Inline(selector.Split(1, buttons)...)
	return c.Send(util.EscapeText(text), &telebot.SendOptions{
		ReplyMarkup: selector,
		ParseMode:   telebot.ModeMarkdownV2,
	})
}

func (h *SimSlotHandler) handleActiveSimSlot(c telebot.Context) error {
	slot, err := strconv.ParseUint(c.Data(), 10, 32)
	if err != nil {
		return err
	}
	if err := h.modem.SetPrimarySimSlot(uint32(slot)); err != nil {
		return err
	}
	return c.Send(fmt.Sprintf("Primary SIM slot set to %d", slot))
}
