package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type ChipHandler struct {
	modemHandler
	data map[int64]string
}

func NewChipHandler(dispatcher *ext.Dispatcher) CommandHandler {
	h := &ChipHandler{
		data: make(map[int64]string, 1),
	}
	h.dispathcer = dispatcher
	h.requiredEuicc = true
	h.next = h.nextHandle
	return h
}

func (h *ChipHandler) Command() string {
	return "chip"
}

func (h *ChipHandler) Description() string {
	return "Get the eUICC chip information"
}

func (h *ChipHandler) nextHandle(b *gotgbot.Bot, ctx *ext.Context) error {
	modem, err := h.modem(ctx)
	if err != nil {
		return err
	}
	modem.Lock()
	defer modem.Unlock()
	usbDevice, err := h.usbDevice(ctx)
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	chip, err := lpac.NewCmd(timeoutCtx, usbDevice).Info()
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
	country, manufacturer, productName := util.FindeUICCManifest(chip.EID)
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
	manufacturerReplacement += " " + chip.EUICCInfo2.SasAccreditationNumber
	manufacturerReplacement = strings.TrimRight(strings.TrimLeft(manufacturerReplacement, " "), " ")

	var keysReplacement string
	for _, key := range chip.EUICCInfo2.PkiForSigning {
		keysReplacement += util.FincCertificateIssuer(key) + "\n"
	}
	keysReplacement = strings.TrimSuffix(keysReplacement, "\n")

	_, err = b.SendMessage(ctx.EffectiveChat.Id,
		util.EscapeText(fmt.Sprintf(message, chip.EID, manufacturerReplacement, chip.EUICCInfo2.ExtCardResource.FreeNonVolatileMemory/1024, keysReplacement)),
		&gotgbot.SendMessageOpts{
			ParseMode: gotgbot.ParseModeMarkdownV2,
		},
	)
	return err
}
