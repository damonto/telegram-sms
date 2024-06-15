package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type ModemHandler struct{}

func NewModemHandler() CommandHandler {
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
EID: %s
`
	for _, m := range modems {
		manufacturer, _ := m.GetManufacturer()
		model, _ := m.GetModel()
		revision, _ := m.GetRevision()
		imei, _ := m.GetImei()
		signal, _ := m.GetSignalQuality()
		network, _ := m.GetOperatorName()
		ICCID, _ := m.GetICCID()

		var eid string
		if m.IsEuicc {
			usbDevice, err := m.GetAtPort()
			if err != nil {
				slog.Error("failed to get AT port", "error", err)
			}
			m.Lock()
			info, err := lpac.NewCmd(context.Background(), usbDevice).Info()
			m.Unlock()
			if err != nil {
				slog.Error("failed to get eSIM info", "error", err)
			} else {
				eid = info.EID
			}
		}

		_, err := b.SendMessage(ctx.EffectiveChat.Id,
			util.EscapeText(fmt.Sprintf(template, manufacturer, model, revision, imei, signal, network, ICCID, eid)), &gotgbot.SendMessageOpts{
				ParseMode: gotgbot.ParseModeMarkdownV2,
			})
		if err != nil {
			return err
		}
	}
	return nil
}
