package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type ProfileHandler struct {
	withModem
	data map[int64]string
}

const (
	ProfileStateHandleAction = "profile_handle_action"
	ProfileStateDelete       = "profile_delete"
	ProfileStateRename       = "profile_rename"

	ProfileActionRename = "rename"
	ProfileActionDelete = "delete"
	ProfileActionEnable = "enable"
)

func NewProfileHandler(dispatcher *ext.Dispatcher) ConversationHandler {
	h := &ProfileHandler{
		data: make(map[int64]string),
	}
	h.dispathcer = dispatcher
	h.next = h.enter
	return h
}

func (h *ProfileHandler) Command() string {
	return "profiles"
}

func (h *ProfileHandler) Description() string {
	return "List all installed eSIM profiles"
}

func (h *ProfileHandler) Conversations() map[string]handlers.Response {
	return map[string]handlers.Response{
		ProfileStateHandleAction: h.handleAction,
		ProfileStateRename:       h.handleActionRename,
		ProfileStateDelete:       h.handleActionDelete,
	}
}

func (h *ProfileHandler) enter(b *gotgbot.Bot, ctx *ext.Context) error {
	modem, err := h.modem(ctx)
	if err != nil {
		return err
	}
	modem.Lock()
	usbDevice, err := h.usbDevice(ctx)
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profiles, err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileList()
	modem.Unlock()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "No profiles found", nil)
		return err
	}
	message, buttons := h.toTextMessage(profiles)
	if _, err := b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(message), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}); err != nil {
		return err
	}

	h.dispathcer.AddHandler(handlers.NewCallback(filters.CallbackQuery(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, "profile_")
	}), func(b *gotgbot.Bot, ctx *ext.Context) error {
		ICCID := strings.TrimPrefix(ctx.CallbackQuery.Data, "profile_")
		_, _, err := b.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
			ChatId:    ctx.EffectiveChat.Id,
			MessageId: ctx.EffectiveMessage.MessageId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{},
			},
		})
		if err != nil {
			return err
		}
		h.data[ctx.EffectiveChat.Id] = ICCID
		return h.handleAskAction(b, ctx)
	}))
	return handlers.NextConversationState(ProfileStateHandleAction)
}

func (h *ProfileHandler) toTextMessage(profiles []lpac.Profile) (string, [][]gotgbot.InlineKeyboardButton) {
	template := `
Name: %s
ICCID: %s
State: *%s*
	`
	var message string
	buttons := make([][]gotgbot.InlineKeyboardButton, 0, len(profiles))
	for _, p := range profiles {
		name := fmt.Sprintf("[%s] ", p.ProviderName)
		if p.Nickname != "" {
			name += p.Nickname
		} else {
			name += p.ProfileName
		}
		message += fmt.Sprintf(template, name, p.ICCID, p.State)
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{
				Text:         fmt.Sprintf("%s (%s)", name, p.ICCID),
				CallbackData: "profile_" + p.ICCID,
			},
		})
	}
	return message, buttons
}

func (h *ProfileHandler) handleAction(b *gotgbot.Bot, ctx *ext.Context) error {
	action := ctx.EffectiveMessage.Text
	switch action {
	case ProfileActionEnable:
		return h.handleActionEnable(b, ctx)
	case ProfileActionRename:
		if _, err := b.SendMessage(ctx.EffectiveChat.Id, "OK. Send me the new name.", nil); err != nil {
			return err
		}
		return handlers.NextConversationState(ProfileStateRename)
	case ProfileActionDelete:
		if _, err := b.SendMessage(ctx.EffectiveChat.Id, "Are you sure you want to delete this profile?", &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
				OneTimeKeyboard: true,
				Keyboard: [][]gotgbot.KeyboardButton{
					{
						{
							Text: "Yes",
						},
						{
							Text: "No",
						},
					},
				},
			},
		}); err != nil {
			return err
		}
		return handlers.NextConversationState(ProfileStateDelete)
	default:
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "Invalid action.", nil)
		return err
	}
}

func (h *ProfileHandler) handleAskAction(b *gotgbot.Bot, ctx *ext.Context) error {
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
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profile, err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileInfo(h.data[ctx.EffectiveChat.Id])
	if err != nil {
		return err
	}
	buttons := []gotgbot.KeyboardButton{
		{
			Text: ProfileActionRename,
		},
	}
	if profile.State == lpac.ProfileStateDisabled {
		buttons = append(buttons, gotgbot.KeyboardButton{
			Text: ProfileActionEnable,
		}, gotgbot.KeyboardButton{
			Text: ProfileActionDelete,
		})
	}

	template := `
You've selected the profile:
ICCID: %s
What do you want to do with this profile?
	`
	_, err = b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(fmt.Sprintf(template, profile.ICCID)), &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.ReplyKeyboardMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			Keyboard:        [][]gotgbot.KeyboardButton{buttons},
		},
	})
	return err
}

func (h *ProfileHandler) handleActionDelete(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.Text != "Yes" {
		if _, err := b.SendMessage(ctx.EffectiveChat.Id, "Profile not deleted.", nil); err != nil {
			return err
		}
		return handlers.EndConversation()
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
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDelete(h.data[ctx.EffectiveChat.Id]); err != nil {
		return err
	}
	delete(h.data, ctx.EffectiveChat.Id)
	if _, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile deleted. /profiles", nil); err != nil {
		return err
	}
	return handlers.EndConversation()
}

func (h *ProfileHandler) handleActionEnable(b *gotgbot.Bot, ctx *ext.Context) error {
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
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileEnable(h.data[ctx.EffectiveChat.Id]); err != nil {
		return err
	}
	delete(h.data, ctx.EffectiveChat.Id)
	if err := modem.Restart(); err != nil {
		return err
	}
	if _, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile enabled. /profiles", nil); err != nil {
		return err
	}
	return handlers.EndConversation()
}

func (h *ProfileHandler) handleActionRename(b *gotgbot.Bot, ctx *ext.Context) error {
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
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileSetNickname(h.data[ctx.EffectiveChat.Id], ctx.EffectiveMessage.Text); err != nil {
		return err
	}
	delete(h.data, ctx.EffectiveChat.Id)
	if _, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile renamed. /profiles", nil); err != nil {
		return err
	}
	return handlers.EndConversation()
}
