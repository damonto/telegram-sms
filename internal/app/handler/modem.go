package handler

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type ModemHandler struct{}

func NewModemHandler() Handler {
	return &ModemHandler{}
}

func (h *ModemHandler) Command() string {
	return "modems"
}

func (h *ModemHandler) Description() string {
	return "List all modems"
}

func (h *ModemHandler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "No modems found", nil)
		return err
	}

	template := `
Manufaturer: %s
Model: %s
Revision: %s
IMEI: %s
Signal: %d
Network: %s
ICCID: %s
`
	for _, m := range modems {
		manufacturer, _ := m.GetManufacturer()
		model, _ := m.GetModel()
		revision, _ := m.GetRevision()
		imei, _ := m.GetImei()
		signal, _ := m.GetSignalQuality()
		network, _ := m.GetOperatorName()
		ICCID, _ := m.GetICCID()

		_, err := b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(fmt.Sprintf(template, manufacturer, model, revision, imei, signal, network, ICCID)), &gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeMarkdownV2,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
