package handler

import (
	"fmt"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

func HandleModemsCommand(c telebot.Context) error {
	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		return c.Send("No modems found.")
	}

	template := `
Manufaturer: %s
Model: %s
Revision: %s
IMEI: %s
Network: %s
Operator: %s
Numbers: %s
Signal: %d%%
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
		operator, _ := m.GetOperatorName()
		operatorCode, _ := m.GetOperatorCode()
		iccid, _ := m.GetIccid()
		numbers, _ := m.GetOwnNumbers()

		message += fmt.Sprintf(
			template,
			util.EscapeText(manufacturer),
			util.EscapeText(model),
			util.EscapeText(revision),
			fmt.Sprintf("`%s`", imei),
			util.EscapeText(util.LookupCarrierName(operatorCode)),
			util.EscapeText(operator),
			util.EscapeText(strings.Join(numbers, ", ")),
			signal,
			fmt.Sprintf("`%s`", iccid),
			fmt.Sprintf("`%s`", m.Eid))
	}
	return c.Send(message, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
}
