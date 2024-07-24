package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

func HandleModemsCommand(c telebot.Context) error {
	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		return c.Send("No modems found")
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
	var message string
	for _, m := range modems {
		manufacturer, _ := m.GetManufacturer()
		model, _ := m.GetModel()
		revision, _ := m.GetRevision()
		imei, _ := m.GetImei()
		signal, _ := m.GetSignalQuality()
		network, _ := m.GetOperatorName()
		ICCID, _ := m.GetICCID()

		var EID string
		if m.IsEuicc {
			var usbDevice string
			var err error
			if config.C.APDUDriver == config.APDUDriverAT {
				usbDevice, err = m.GetAtPort()
			} else {
				usbDevice, err = m.GetQMIDevice()
			}
			if err != nil {
				slog.Error("failed to get AT port", "error", err)
			}
			m.Lock()
			info, err := lpac.NewCmd(context.Background(), usbDevice).Info()
			m.Unlock()
			if err != nil {
				slog.Error("failed to get eUICC info", "error", err)
			} else {
				EID = info.EID
			}
		}
		message += fmt.Sprintf(
			template,
			manufacturer,
			model,
			revision,
			fmt.Sprintf("`%s`", imei),
			signal,
			network,
			fmt.Sprintf("`%s`", ICCID),
			fmt.Sprintf("`%s`", EID)) + "\n"
	}
	return c.Send(util.EscapeText(strings.TrimRight(message, "\n")), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
}
