package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"gopkg.in/telebot.v3"
)

type DownloadHandler struct {
	handler
	activationCode       *lpac.ActivationCode
	downloadConfirmation chan bool
}

const (
	StateDownloadAskActivationCode   = "download_ask_activation_code"
	StateDownloadAskConfirmationCode = "download_ask_confirmation_code"
)

func HandleDownloadCommand(c telebot.Context) error {
	h := &DownloadHandler{
		downloadConfirmation: make(chan bool, 1),
	}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.States(map[string]telebot.HandlerFunc{
		StateDownloadAskActivationCode:   h.handleActivationCode,
		StateDownloadAskConfirmationCode: h.handleConfirmationCode,
	})
	return h.handle(c)
}

func (h *DownloadHandler) handle(c telebot.Context) error {
	h.state.Next(StateDownloadAskActivationCode)
	return c.Send("Please send me the activation code.")
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
		return c.Send("Please send me the confirmation code.")
	}

	h.stateManager.Done(c)
	return h.download(c)
}

func (h *DownloadHandler) handleConfirmationCode(c telebot.Context) error {
	confirmationCode := c.Text()
	if confirmationCode == "" {
		h.state.Next(StateDownloadAskConfirmationCode)
		return c.Send("Invalid confirmation code.")
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
	usbDevice, err := h.GetUsbDevice()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDownload(h.activationCode, func(current string, profileMetadata *lpac.Profile, downloadConfirmation chan bool) error {
		return h.handleDownloadProgress(c, message, current, profileMetadata, downloadConfirmation)
	}); err != nil {
		if errors.Is(err, lpac.ErrDownloadCancelled) {
			_, err := c.Bot().Edit(message, "Download cancelled. /profiles")
			return err
		}
		slog.Info("failed to download profile", "error", err)
		_, err := c.Bot().Edit(message, "Failed to download profile: "+err.Error())
		return err
	}
	_, err = c.Bot().Edit(message, "Congratulations! Your profile has been downloaded. /profiles")
	return err
}

func (h *DownloadHandler) handleDownloadProgress(c telebot.Context, message *telebot.Message, current string, profileMetadata *lpac.Profile, downloadConfirmation chan bool) error {
	if profileMetadata != nil && current == lpac.ProgressMetadataParse {
		template := `
Are you sure you want to download this profile?
Provider Name: %s
Profile Name: %s
ICCID: %s
`
		selector := telebot.ReplyMarkup{}
		btns := make([]telebot.Btn, 0, 2)
		for _, action := range []string{"Yes", "No"} {
			btn := selector.Data(action, fmt.Sprint(time.Now().UnixNano()), action)
			c.Bot().Handle(&btn, func(c telebot.Context) error {
				h.downloadConfirmation <- c.Callback().Data == "Yes"
				return nil
			})
			btns = append(btns, btn)
		}
		selector.Inline(btns)
		_, err := c.Bot().Edit(message, fmt.Sprintf(template, profileMetadata.ProviderName, profileMetadata.ProfileName, profileMetadata.ICCID), &selector)
		return err
	}

	if current == lpac.ProgressPreviewConfirm {
		downloadConfirmation <- <-h.downloadConfirmation
		return nil
	}

	_, err := c.Bot().Edit(message, current)
	return err
}
