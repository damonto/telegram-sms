package handler

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
)

type Handler interface {
	Command() string
	Description() string
	Handle(bot *gotgbot.Bot, ctx *ext.Context) error
}

type ConversationHandler interface {
	Handler
	Conversations() map[string]handlers.Response
}

type withModem struct {
	dispathcer *ext.Dispatcher
	modemId    string
	next       handlers.Response
}

var (
	ErrNextHandlerNotSet = fmt.Errorf("next handler not set")
)

func (h *withModem) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	if h.next == nil {
		return ErrNextHandlerNotSet
	}

	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		return modem.ErrModemNotFound
	}

	if len(modems) == 1 {
		for _, m := range modems {
			imei, _ := m.GetImei()
			h.modemId = imei
		}
		slog.Info("only one modem found, using it", "modemId", h.modemId)
		return h.next(b, ctx)
	}
	done := make(chan struct{}, 1)
	if err := h.selectModem(modems, done, b, ctx); err != nil {
		return err
	}
	<-done
	return h.next(b, ctx)
}

func (h *withModem) selectModem(modems map[string]*modem.Modem, done chan struct{}, b *gotgbot.Bot, ctx *ext.Context) error {
	buttons := make([][]gotgbot.InlineKeyboardButton, 0, len(modems))
	for _, m := range modems {
		imei, _ := m.GetImei()
		model, _ := m.GetModel()
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s (%s)", model, imei),
				CallbackData: "modem_" + imei,
			},
		})
	}

	h.dispathcer.AddHandler(handlers.NewCallback(filters.CallbackQuery(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, "modem_")
	}), func(b *gotgbot.Bot, ctx *ext.Context) error {
		h.modemId = strings.TrimPrefix(ctx.CallbackQuery.Data, "modem_")
		done <- struct{}{}
		return nil
	}))

	_, err := b.SendMessage(ctx.EffectiveChat.Id, "I found the following modems, please select one:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	return err
}
