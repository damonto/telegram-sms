package handler

import (
	"fmt"

	"github.com/damonto/euicc-go/lpa"
	"github.com/damonto/telegram-sms/internal/app/state"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type DownloadHandler struct {
	*Handler
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
	return h
}

func (h *DownloadHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		fmt.Println("DownloadHandler Handle")
		return nil
	}
}

func (h *DownloadHandler) HandleMessage(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	return nil
}

func (h *DownloadHandler) HandleCallbackQuery(ctx *th.Context, query telego.CallbackQuery, s *state.ChatState) error {
	return nil
}
