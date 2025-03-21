package handler

import (
	"fmt"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type ChipHandler struct {
	*Handler
}

const ChipMessageTemplate = `
EID: %s
Brand: %s
Manufacturer: %s
Free Space: %d KiB
Sign Keys:
%s
`

func NewChipHandler() *ChipHandler {
	h := new(ChipHandler)
	return h
}

func (h *ChipHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		lpa, err := h.LPA(ctx)
		if err != nil {
			return err
		}
		defer lpa.Close()
		info, err := lpa.Info()
		if err != nil {
			return err
		}
		_, err = h.Reply(ctx, update, h.message(info), nil)
		return err
	}
}

func (h *ChipHandler) message(info *lpa.Info) string {
	manufacturer := h.manufacturer(info)
	var key string
	for _, k := range info.Certificates {
		key += k + "\n"
	}
	return fmt.Sprintf(
		ChipMessageTemplate,
		fmt.Sprintf("`%s`", info.EID),
		util.EscapeText(manufacturer),
		util.EscapeText(util.If(info.Manufacturer != "", info.Manufacturer, info.SasAcreditationNumber)),
		info.FreeSpace/1024,
		util.EscapeText(strings.TrimRight(key, "\n")),
	)
}

func (h *ChipHandler) manufacturer(info *lpa.Info) string {
	var manufacturer string
	if info.Product.Country != "" {
		manufacturer += string(0x1F1E6+rune(info.Product.Country[0])-'A') + string(0x1F1E6+rune(info.Product.Country[1])-'A')
	}
	if info.Product.Manufacturer != "" {
		manufacturer += " " + info.Product.Manufacturer
	}
	if info.Product.Brand != "" {
		manufacturer += " " + info.Product.Brand
	}
	return strings.TrimRight(strings.TrimLeft(manufacturer, " "), " ")
}
