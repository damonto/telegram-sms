package handler

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type DownloadHandler struct {
	data map[int64]*lpac.ActivationCode
}

const (
	DownloadStateActivationCode   = "activation_code"
	DownloadStateConfirmationCode = "confirmation_code"
)

func NewDownloadHandler() ConversationHandler {
	return &DownloadHandler{
		data: make(map[int64]*lpac.ActivationCode, 1),
	}
}

func (h *DownloadHandler) Command() string {
	return "download"
}

func (h *DownloadHandler) Description() string {
	return "Download an eSIM profile"
}

func (h *DownloadHandler) Handle(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Please send me the activation code or QR Code", nil)
	if err != nil {
		return err
	}
	return handlers.NextConversationState(DownloadStateActivationCode)
}

func (h *DownloadHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		DownloadStateActivationCode:   h.handleActivationCode,
		DownloadStateConfirmationCode: h.handleConfirmationCode,
	}
}

func (h *DownloadHandler) handleActivationCode(b *gotgbot.Bot, ctx *ext.Context) error {
	activationCode := ctx.EffectiveMessage.Text
	if activationCode == "" || !strings.HasPrefix(activationCode, "LPA:1$") {
		_, err := ctx.EffectiveMessage.Reply(b, "Invalid activation code", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateActivationCode)
	}

	parts := strings.Split(activationCode, "$")
	activationCodeStruct := &lpac.ActivationCode{
		SMDP:       parts[1],
		MatchingId: parts[2],
	}
	if len(parts) == 5 && parts[4] == "1" {
		h.data[ctx.EffectiveMessage.Chat.Id] = activationCodeStruct
		_, err := ctx.EffectiveMessage.Reply(b, "Please send me the confirmation code", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateConfirmationCode)
	}

	if err := h.download(b, ctx, activationCodeStruct); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Failed to download profile", nil)
		if err != nil {
			return err
		}
		return handlers.EndConversation()
	}
	return handlers.EndConversation()
}

func (h *DownloadHandler) handleConfirmationCode(b *gotgbot.Bot, ctx *ext.Context) error {
	confirmationCode := ctx.EffectiveMessage.Text
	if confirmationCode == "" {
		_, err := ctx.EffectiveMessage.Reply(b, "Invalid confirmation code", nil)
		if err != nil {
			return err
		}
		return handlers.NextConversationState(DownloadStateConfirmationCode)
	}

	activationCode := h.data[ctx.EffectiveMessage.Chat.Id]
	delete(h.data, ctx.EffectiveMessage.Chat.Id)
	activationCode.ConfirmationCode = confirmationCode

	if err := h.download(b, ctx, activationCode); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, "Failed to download profile", nil)
		if err != nil {
			return err
		}
		return handlers.EndConversation()
	}
	return handlers.EndConversation()
}

func (h *DownloadHandler) download(b *gotgbot.Bot, ctx *ext.Context, activationCode *lpac.ActivationCode) error {
	text := "Downloading profile..."
	message, err := b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(text), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
	})
	if err != nil {
		return err
	}
	fmt.Println(message.MessageId)
	return nil
}
