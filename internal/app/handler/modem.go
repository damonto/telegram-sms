package handler

import (
	"fmt"
	"log/slog"

	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func ListModem(mm *modem.Manager) th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		modems, err := mm.Modems()
		if err != nil {
			return err
		}

		if len(modems) == 0 {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(update.Message.Chat.ID), "No modems found"))
			return err
		}

		template := `
Manufaturer: %s
Model: %s
Revision: %s
IMEI: %s
Network: %s
Operator: %s
Number: %s
Signal: %d%%
ICCID: %s
EID: %s
`
		var message string
		for _, m := range modems {
			percent, _, _ := m.SignalQuality()
			lpa, err := lpa.NewLPA(m)
			if err != nil {
				slog.Error("Failed to create LPA", "error", err)
			}
			info, _ := lpa.Info()
			var eid string
			if info != nil {
				eid = info.EID
			}
			code, _ := m.OperatorCode()
			message += fmt.Sprintf(template,
				util.EscapeText(m.Manufacturer),
				util.EscapeText(m.Model),
				util.EscapeText(m.FirmwareRevision),
				m.EquipmentIdentifier,
				util.EscapeText(util.LookupCarrier(code)),
				util.EscapeText(util.LookupCarrier(m.Sim.OperatorIdentifier)),
				util.EscapeText(m.Number),
				percent,
				m.Sim.Identifier,
				eid)
		}
		_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(update.Message.Chat.ID), message).WithParseMode(telego.ModeMarkdownV2))
		return err
	}
}
