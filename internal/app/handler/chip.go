package handler

import (
	"fmt"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type ChipHandler struct {
	handler
}

func HandleChipCommand(c telebot.Context) error {
	h := &ChipHandler{}
	h.init(c)
	return h.handle(c)
}

func (h *ChipHandler) handle(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()
	lpa, err := h.GetLPA()
	if err != nil {
		return err
	}

	message := `
EID: %s
Manufaturer: %s
Free Space: %d KiB
Sign Keys:
%s
`
	eid, err := lpa.GetEid()
	if err != nil {
		return err
	}
	chip, err := lpa.GetEuiccInfo2()
	if err != nil {
		return err
	}
	defer lpa.Close()
	country, manufacturer, productName := util.LookupEUM(eid)
	var manufacturerReplacement string
	if country != "" {
		manufacturerReplacement += string(0x1F1E6+rune(country[0])-'A') + string(0x1F1E6+rune(country[1])-'A')
	}
	if manufacturer != "" {
		manufacturerReplacement += " " + manufacturer
	}
	if productName != "" {
		manufacturerReplacement += " " + productName
	}
	manufacturerReplacement += " " + chip.SasAccreditationNumber
	manufacturerReplacement = strings.TrimRight(strings.TrimLeft(manufacturerReplacement, " "), " ")

	var keysReplacement string
	for _, key := range chip.CiPKIdForSigning {
		keysReplacement += util.FindCertificateIssuer(key) + "\n"
	}
	keysReplacement = strings.TrimSuffix(keysReplacement, "\n")
	return c.Send(
		fmt.Sprintf(
			message,
			fmt.Sprintf("`%s`", eid),
			util.EscapeText(manufacturerReplacement),
			chip.ExtCardResource.FreeNonVolatileMemory/1024,
			util.EscapeText(keysReplacement),
		),
		&telebot.SendOptions{
			ParseMode: telebot.ModeMarkdownV2,
		},
	)
}
