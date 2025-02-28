package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

const CallbackQueryAskModemPrefix = "ask_modem"

type ModemRequiredMiddleware struct {
	mm    *modem.Manager
	modem chan *modem.Modem
}

func NewModemRequiredMiddleware(mm *modem.Manager, handler *th.BotHandler) *ModemRequiredMiddleware {
	m := &ModemRequiredMiddleware{
		mm:    mm,
		modem: make(chan *modem.Modem, 1),
	}
	handler.HandleCallbackQuery(m.HandleModemSelectionCallbackQuery, th.CallbackDataPrefix(CallbackQueryAskModemPrefix))
	return m
}

func (m *ModemRequiredMiddleware) Middleware(eUICCRequired bool) th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		modems, err := m.mm.Modems()
		if err != nil {
			return err
		}
		if len(modems) == 0 {
			return m.sendErrorModemNotFound(ctx, update)
		}
		if eUICCRequired {
			for path, modem := range modems {
				// lpa.New will open the ISD-R logical channel, if it fails, the modem is not an eUICC.
				l, err := lpa.New(modem)
				if err != nil {
					delete(modems, path)
					slog.Error("Failed to create LPA", "error", err)
				}
				slog.Debug("The SIM card is an eUICC", "objectPath", path)
				l.Close()
			}
		}
		return m.run(modems, ctx, update)
	}
}

func (m *ModemRequiredMiddleware) run(modems map[dbus.ObjectPath]*modem.Modem, ctx *th.Context, update telego.Update) error {
	if len(modems) == 0 {
		return m.sendErrorModemNotFound(ctx, update)
	}
	// If there is only one modem, select it automatically.
	if len(modems) == 1 {
		for _, modem := range modems {
			m.modem <- modem
			ctx = ctx.WithValue("modem", modem)
			return ctx.Next(update)
		}
	}
	if err := m.ask(ctx, update, modems); err != nil {
		return err
	}
	modem := <-m.modem
	ctx = ctx.WithValue("modem", modem)
	return ctx.Next(update)
}

func (m *ModemRequiredMiddleware) HandleModemSelectionCallbackQuery(ctx *th.Context, query telego.CallbackQuery) error {
	objectPath := query.Data[len(CallbackQueryAskModemPrefix)+1:]
	slog.Info("Modem selected", "objectPath", objectPath)
	modems, err := m.mm.Modems()
	if err != nil {
		return err
	}
	m.modem <- modems[dbus.ObjectPath(objectPath)]
	if err := ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
	}); err != nil {
		return err
	}
	return ctx.Bot().DeleteMessage(ctx, &telego.DeleteMessageParams{
		ChatID:    tu.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
	})
}

func (m *ModemRequiredMiddleware) sendErrorModemNotFound(ctx *th.Context, update telego.Update) error {
	_, err := ctx.Bot().SendMessage(
		ctx,
		tu.Message(
			tu.ID(update.Message.From.ID),
			"No modems found. Please connect a modem and try again.",
		).WithReplyParameters(&telego.ReplyParameters{
			MessageID: update.Message.MessageID,
		}),
	)
	if err != nil {
		return err
	}
	return errors.New("no modem found")
}

func (m *ModemRequiredMiddleware) ask(ctx *th.Context, update telego.Update, modems map[dbus.ObjectPath]*modem.Modem) error {
	var err error
	var buttons [][]telego.InlineKeyboardButton
	for path, modem := range modems {
		buttons = append(buttons, tu.InlineKeyboardRow(telego.InlineKeyboardButton{
			Text:         modem.Model,
			CallbackData: fmt.Sprintf("%s:%s", CallbackQueryAskModemPrefix, path),
		}))
	}
	var message string
	for _, modem := range modems {
		message += fmt.Sprintf(`
*%s*
Manufacturer: %s
IMEI: %s
Firmware revision: %s
		`, util.EscapeText(modem.Model),
			util.EscapeText(modem.Manufacturer),
			modem.EquipmentIdentifier,
			util.EscapeText(modem.FirmwareRevision))
	}
	_, err = ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(update.Message.From.ID),
		strings.TrimRight(message, "\n"),
	).WithReplyMarkup(tu.InlineKeyboard(buttons...)).WithReplyParameters(&telego.ReplyParameters{
		MessageID: update.Message.MessageID,
	}).WithParseMode(telego.ModeMarkdownV2))
	return err
}
