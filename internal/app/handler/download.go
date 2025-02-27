package handler

import (
	"net/url"
	"strings"

	"github.com/damonto/euicc-go/lpa"
	"github.com/damonto/telegram-sms/internal/app/state"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type DownloadHandler struct {
	*Handler
	cc chan string
}

type DownloadValue struct {
	Modem         *modem.Modem
	ActvationCode *lpa.ActivationCode
}

const (
	DownloadAskActivationCode   state.State = "download_ask_activation_code"
	DownloadAskConfirmationCode state.State = "download_ask_confirmation_code"
	DownloadConfirm             state.State = "download_confirm"
)

func NewDownloadHandler() state.Handler {
	h := new(DownloadHandler)
	h.cc = make(chan string, 1)
	return h
}

func (h *DownloadHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		m := h.Modem(ctx)
		state.M.Enter(update.Message.Chat.ID, &state.ChatState{
			Handler: h,
			State:   DownloadAskActivationCode,
			Value:   &DownloadValue{Modem: m},
		})
		_, err := h.Reply(ctx, update, "Please enter the activation code", nil)
		return err
	}
}

func (h *DownloadHandler) HandleMessage(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	value := s.Value.(*DownloadValue)
	if s.State == DownloadAskActivationCode {
		return h.downloadProfile(ctx, message, s, value)
	}
	return nil
}

func (h *DownloadHandler) downloadProfile(ctx *th.Context, message telego.Message, s *state.ChatState, value *DownloadValue) error {
	ac, ccRequired, err := h.parseActivationCode(value, message.Text)
	if err != nil {
		return err
	}
	value.ActvationCode = ac
	if ccRequired {
		state.M.Current(message.From.ID, DownloadAskConfirmationCode)
		_, err := h.ReplyMessage(ctx, message, util.EscapeText("Please enter the confirmation code"), nil)
		return err
	}
	return h.download(ctx, message, s, value)
}

func (h *DownloadHandler) download(ctx *th.Context, message telego.Message, s *state.ChatState, value *DownloadValue) error {
	return nil
}

// LPA:1$
func (h *DownloadHandler) parseActivationCode(value *DownloadValue, text string) (ac *lpa.ActivationCode, ccRequired bool, err error) {
	parts := strings.Split(text, "$")
	ac = &lpa.ActivationCode{
		SMDP: &url.URL{Scheme: "https", Host: parts[1]},
		IMEI: value.Modem.EquipmentIdentifier,
	}
	if len(parts) == 3 {
		ac.MatchingID = parts[2]
	}
	if len(parts) == 5 && parts[4] == "1" {
		ccRequired = true
	}
	return ac, ccRequired, nil
}

func (h *DownloadHandler) HandleCallbackQuery(ctx *th.Context, query telego.CallbackQuery, s *state.ChatState) error {
	return nil
}
