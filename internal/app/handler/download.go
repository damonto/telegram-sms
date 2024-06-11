package handler

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type DownloadHandler struct {
	withModem
	data map[int64]lpac.ActivationCode
}

const (
	DownloadStateActivationCode   = "esim_download_activation_code"
	DownloadStateConfirmationCode = "esim_download_confirmation_code"
)

func NewDownloadHandler(dispatcher *ext.Dispatcher) ConversationHandler {
	h := &DownloadHandler{
		data: make(map[int64]lpac.ActivationCode, 1),
	}
	h.dispathcer = dispatcher
	h.next = h.enter
	return h
}

func (h *DownloadHandler) Command() string {
	return "download"
}

func (h *DownloadHandler) Description() string {
	return "Download an eSIM profile"
}

func (h *DownloadHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		DownloadStateActivationCode:   h.handleActivationCode,
		DownloadStateConfirmationCode: h.handleConfirmationCode,
	}
}

func (h *DownloadHandler) enter(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, "Please send me the activation code.", nil)
	if err != nil {
		return err
	}
	return handlers.NextConversationState(DownloadStateActivationCode)
}

func (h *DownloadHandler) handleActivationCode(b *gotgbot.Bot, ctx *ext.Context) error {
	activationCode := ctx.EffectiveMessage.Text
	if activationCode == "" || !strings.HasPrefix(activationCode, "LPA:1$") {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "Invalid activation code.", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateActivationCode)
	}

	parts := strings.Split(activationCode, "$")
	activationCodeStruct := lpac.ActivationCode{
		SMDP:       parts[1],
		MatchingId: parts[2],
	}
	if len(parts) == 5 && parts[4] == "1" {
		h.data[ctx.EffectiveMessage.Chat.Id] = activationCodeStruct
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "Please send me the confirmation code", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateConfirmationCode)
	}

	if err := h.download(b, ctx, activationCodeStruct); err != nil {
		handlers.EndConversation()
		return err
	}
	return handlers.EndConversation()
}

func (h *DownloadHandler) handleConfirmationCode(b *gotgbot.Bot, ctx *ext.Context) error {
	confirmationCode := ctx.EffectiveMessage.Text
	if confirmationCode == "" {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "Invalid confirmation code", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateConfirmationCode)
	}

	activationCode := h.data[ctx.EffectiveMessage.Chat.Id]
	delete(h.data, ctx.EffectiveMessage.Chat.Id)
	activationCode.ConfirmationCode = confirmationCode
	if err := h.download(b, ctx, activationCode); err != nil {
		handlers.EndConversation()
		return err
	}
	return handlers.EndConversation()
}

func (h *DownloadHandler) download(b *gotgbot.Bot, ctx *ext.Context, activationCode lpac.ActivationCode) error {
	text := "Downloading..."
	message, err := b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(text), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	})
	if err != nil {
		return err
	}
	modem, err := h.modem(ctx)
	if err != nil {
		return err
	}
	modem.Lock()
	defer modem.Unlock()

	usbDevice, err := h.usbDevice(ctx)
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDownload(activationCode, func(current string) error {
		text := "Downloading... \n" + current
		_, _, err := message.EditText(b, text, nil)
		return err
	}); err != nil {
		slog.Info("failed to download profile", "error", err)
		message.EditText(b, "Failed to download profile\n"+err.Error(), nil)
		return err
	}
	_, _, err = message.EditText(b, "Congratulations! Your profile has been downloaded. /profiles", nil)
	return err
}
