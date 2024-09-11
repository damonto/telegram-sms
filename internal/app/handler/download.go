package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/damonto/libeuicc-go"
	"gopkg.in/telebot.v3"
)

type DownloadHandler struct {
	handler
	activationCode       *libeuicc.ActivationCode
	confirmationCode     chan string
	downloadConfirmation chan bool
}

const (
	StateDownloadAskActivationCode             = "download_ask_activation_code"
	StateDownloadAskConfirmationCode           = "download_ask_confirmation_code"
	StateDownloadAskConfirmationCodeInDownload = "download_ask_confirmation_code_in_download"
)

func HandleDownloadCommand(c telebot.Context) error {
	h := &DownloadHandler{
		downloadConfirmation: make(chan bool, 1),
		confirmationCode:     make(chan string, 1),
	}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.States(map[string]telebot.HandlerFunc{
		StateDownloadAskActivationCode:             h.handleAskActivationCode,
		StateDownloadAskConfirmationCode:           h.handleAskConfirmationCode,
		StateDownloadAskConfirmationCodeInDownload: h.handleAskConfirmationCodeInDownload,
	})
	return h.handle(c)
}

func (h *DownloadHandler) handle(c telebot.Context) error {
	h.state.Next(StateDownloadAskActivationCode)
	return c.Send("Please send me the activation code.")
}

func (h *DownloadHandler) handleAskActivationCode(c telebot.Context) error {
	activationCode := c.Text()
	if activationCode == "" || !strings.HasPrefix(activationCode, "LPA:1$") {
		h.state.Next(StateDownloadAskActivationCode)
		return c.Send("Invalid activation code.")
	}

	parts := strings.Split(activationCode, "$")
	h.activationCode = &libeuicc.ActivationCode{
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

func (h *DownloadHandler) handleAskConfirmationCode(c telebot.Context) error {
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

	l, err := h.GetLPA()
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	err = l.Download(timeoutCtx, h.activationCode, &libeuicc.DownloadOption{
		ProgressBar: func(progress libeuicc.DownloadProgress) {
			if progress == libeuicc.DownloadProgressConfirmDownload || progress == libeuicc.DownloadProgressConfirmationCodeRequired {
				return
			}
			progressBar := strings.Repeat("⣿", int(progress)) + strings.Repeat("⣀", 10-int(progress-1))
			percent := (progress - 1) * 10
			if _, err := c.Bot().Edit(message, fmt.Sprintf("⏳ Downloading\n%s %d%% \n This may take a few minutes.", progressBar, percent)); err != nil {
				slog.Error("failed to edit message", "error", err)
				cancel()
			}
		},
		ConfirmFunc: func(metadata *libeuicc.ProfileMetadata) bool {
			return h.handleConfirmDownload(c, metadata)
		},
		ConfirmationCodeFunc: func() string {
			h.state.Next(StateDownloadAskConfirmationCodeInDownload)
			return <-h.confirmationCode
		},
	})
	defer l.Close()

	if err != nil {
		if err == libeuicc.ErrDownloadCanceled {
			c.Bot().Edit(message, "Download canceled.")
			return nil
		}
		slog.Error("failed to download profile", "error", err)
		c.Bot().Edit(message, "Failed to download profile. Error: "+err.Error())
		return err
	}

	_, err = c.Bot().Edit(message, "Congratulations! Your profile has been downloaded. /profiles")
	return err
}

func (h *DownloadHandler) handleAskConfirmationCodeInDownload(c telebot.Context) error {
	confirmationCode := c.Text()
	if confirmationCode == "" {
		h.state.Next(StateDownloadAskConfirmationCodeInDownload)
		return c.Send("Invalid confirmation code.")
	}
	h.confirmationCode <- confirmationCode
	return nil
}

func (h *DownloadHandler) handleConfirmDownload(c telebot.Context, metadata *libeuicc.ProfileMetadata) bool {
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

	confirmMessage, err := c.Bot().Send(c.Recipient(), fmt.Sprintf(template, metadata.ProviderName, metadata.ProfileName, metadata.ICCID), &selector)
	if err != nil {
		slog.Error("failed to send profile metadata", "error", err)
		return false
	}
	defer c.Bot().Delete(confirmMessage)
	return <-h.downloadConfirmation
}
