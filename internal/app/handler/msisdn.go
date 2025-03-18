package handler

import (
	"log/slog"

	"github.com/damonto/telegram-sms/internal/app/state"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type MSISDNHandler struct {
	*Handler
}

type MSISDNValue struct {
	Modem *modem.Modem
}

func NewMSISDNHandler() state.Handler {
	h := new(MSISDNHandler)
	return h
}

func (h *MSISDNHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		state.M.Enter(update.Message.Chat.ID, &state.ChatState{
			Handler: h,
			Value:   &MSISDNValue{Modem: h.Modem(ctx)},
		})
		_, err := h.Reply(ctx, update, util.EscapeText("Okay, Send me the new MSISDN you want to update."), nil)
		return err
	}
}

func (h *MSISDNHandler) HandleMessage(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	value := s.Value.(*MSISDNValue)
	name := util.If(
		value.Modem.Sim.OperatorName != "",
		value.Modem.Sim.OperatorName,
		util.LookupCarrier(value.Modem.Sim.OperatorIdentifier),
	)
	if err := value.Modem.SetMSISDN(name, message.Text); err != nil {
		h.ReplyMessage(ctx, message, util.EscapeText(err.Error()), nil)
		return err
	}
	if err := value.Modem.Restart(); err != nil {
		slog.Warn("Failed to restart modem", "error", err)
	}
	h.ReplyMessage(ctx, message, util.EscapeText("I have updated the MSISDN on the SIM. If you don't see the changes, you may need to restart the ModemManager manually."), nil)
	return nil
}

func (h *MSISDNHandler) HandleCallbackQuery(ctx *th.Context, query telego.CallbackQuery, s *state.ChatState) error {
	return nil
}
