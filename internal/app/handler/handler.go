package handler

import (
	"fmt"
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
	next       handlers.Response
	notifier   map[int64]chan string
	modems     map[int64]*modem.Modem
}

var (
	ErrNextHandlerNotSet = fmt.Errorf("next handler not set")
)

func (h *withModem) Init() {
	if h.notifier == nil {
		h.notifier = make(map[int64]chan string)
	}
	if h.modems == nil {

		h.modems = make(map[int64]*modem.Modem)
	}
}

func (h *withModem) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	if h.next == nil {
		return ErrNextHandlerNotSet
	}
	h.Init()
	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		return modem.ErrModemNotFound
	}
	if len(modems) == 1 {
		for _, m := range modems {
			h.modems[ctx.EffectiveChat.Id] = m
		}
		return h.next(b, ctx)
	}

	h.notifier[ctx.EffectiveChat.Id] = make(chan string, 1)
	if err := h.selectModem(modems, b, ctx); err != nil {
		return err
	}
	modem, err := modem.GetManager().GetModem(<-h.notifier[ctx.EffectiveChat.Id])
	if err != nil {
		return err
	}
	h.modems[ctx.EffectiveChat.Id] = modem
	return h.next(b, ctx)
}

func (h *withModem) selectModem(modems map[string]*modem.Modem, b *gotgbot.Bot, ctx *ext.Context) error {
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
		h.notifier[ctx.EffectiveChat.Id] <- strings.TrimPrefix(ctx.CallbackQuery.Data, "modem_")
		_, err := b.DeleteMessage(ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, nil)
		return err
	}))

	_, err := b.SendMessage(ctx.EffectiveChat.Id, "I found the following modems, please select one:", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	return err
}

func (h *withModem) modem(ctx *ext.Context) (*modem.Modem, error) {
	m, ok := h.modems[ctx.EffectiveChat.Id]
	if !ok {
		return nil, modem.ErrModemNotFound
	}
	return m, nil
}

func (h *withModem) usbDevice(ctx *ext.Context) (string, error) {
	m, err := h.modem(ctx)
	if err != nil {
		return "", err
	}
	return m.GetAtPort()
}
