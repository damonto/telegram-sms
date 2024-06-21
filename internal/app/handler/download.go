package handler

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type DownloadHandler struct {
	handler
	activationCode *lpac.ActivationCode
}

const (
	StateDownloadAskActivationCode   = "download_ask_activation_code"
	StateDownloadAskConfirmationCode = "download_ask_confirmation_code"
)

func HandleDownloadCommand(c telebot.Context) error {
	h := &DownloadHandler{}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.Stages(map[string]telebot.HandlerFunc{
		StateDownloadAskActivationCode:   h.handleActivationCode,
		StateDownloadAskConfirmationCode: h.handleConfirmationCode,
	})
	return h.handle(c)
}

func (h *DownloadHandler) handle(c telebot.Context) error {
	h.state.Next(StateDownloadAskActivationCode)
	return c.Send("Please send me the activation code")
}

func (h *DownloadHandler) handleActivationCode(c telebot.Context) error {
	activationCode := c.Text()
	if activationCode == "" || !strings.HasPrefix(activationCode, "LPA:1$") {
		h.state.Next(StateDownloadAskActivationCode)
		return c.Send("Invalid activation code.")
	}

	parts := strings.Split(activationCode, "$")
	h.activationCode = &lpac.ActivationCode{
		SMDP:       parts[1],
		MatchingId: parts[2],
	}
	if len(parts) == 5 && parts[4] == "1" {
		h.state.Next(StateDownloadAskConfirmationCode)
		return c.Send("Please send me the confirmation code")
	}

	h.stateManager.Done(c)
	return h.download(c)
}

func (h *DownloadHandler) handleConfirmationCode(c telebot.Context) error {
	confirmationCode := c.Text()
	if confirmationCode == "" {
		h.state.Next(StateDownloadAskConfirmationCode)
		return c.Send("Invalid confirmation code")
	}

	h.activationCode.ConfirmationCode = confirmationCode
	h.stateManager.Done(c)
	return h.download(c)
}

func (h *DownloadHandler) download(c telebot.Context) error {
	message, err := c.Bot().Send(c.Recipient(), "⏳Downloading")
	if err != nil {
		return err
	}

	h.modem.Lock()
	defer h.modem.Unlock()
	usbDevice, err := h.modem.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDownload(h.activationCode, func(current string) error {
		_, err := c.Bot().Edit(message, util.EscapeText("⏳"+current))
		return err
	}); err != nil {
		slog.Info("failed to download profile", "error", err)
		_, err := c.Bot().Edit(message, util.EscapeText("Failed to download profile: "+err.Error()))
		return err
	}
	_, err = c.Bot().Edit(message, "Congratulations! Your profile has been downloaded. /profiles")
	return err
}
