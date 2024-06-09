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
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type ProfileHandler struct {
	withModem
}

const (
	ProfileStateActionRename        = "profile_action_rename"
	ProfileStateActionConfirmDelete = "profile_action_confirm_delete"
)

func NewProfileHandler(dispatcher *ext.Dispatcher) ConversationHandler {
	h := &ProfileHandler{}
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
		ProfileStateActionRename: h.handleActionRename,
		// ProfileStateActionConfirmDelete: h.handleActionConfirmDelete,
	}
}

func (h *ProfileHandler) enter(b *gotgbot.Bot, ctx *ext.Context) error {
	m, err := modem.GetManager().GetModem(h.modemId)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profiles, err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileList()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "No profiles found", nil)
		return err
	}

	h.dispathcer.AddHandler(handlers.NewCallback(filters.CallbackQuery(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, "profile_") && !strings.HasPrefix(cq.Data, "profile_action_")
	}), func(b *gotgbot.Bot, ctx *ext.Context) error {
		ICCID := strings.TrimPrefix(ctx.CallbackQuery.Data, "profile_")
		return h.handleAction(b, ctx, ICCID)
	}))

	message, buttons := h.profileMessage(profiles)
	_, err = b.SendMessage(ctx.EffectiveChat.Id, util.EscapeText(message), &gotgbot.SendMessageOpts{
		ParseMode: gotgbot.ParseModeMarkdownV2,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	return err
}

func (h *ProfileHandler) handleAction(b *gotgbot.Bot, ctx *ext.Context, ICCID string) error {
	m, err := modem.GetManager().GetModem(h.modemId)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profile, err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileInfo(ICCID)
	if err != nil {
		return err
	}
	buttons := []gotgbot.InlineKeyboardButton{
		{
			Text:         "Rename",
			CallbackData: "profile_action_rename_" + ICCID,
		},
	}
	if profile.State == lpac.ProfileStateDisabled {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{
			Text:         "Enable",
			CallbackData: "profile_action_enable_" + ICCID,
		}, gotgbot.InlineKeyboardButton{
			Text:         "Delete",
			CallbackData: "profile_action_delete_" + ICCID,
		})
	}

	h.dispathcer.AddHandler(handlers.NewCallback(filters.CallbackQuery(func(cq *gotgbot.CallbackQuery) bool {
		return strings.HasPrefix(cq.Data, "profile_action_")
	}), func(b *gotgbot.Bot, ctx *ext.Context) error {
		return h.doAction(b, ctx)
	}))

	_, err = b.SendMessage(ctx.EffectiveChat.Id, "What do you want to do with this profile?", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{buttons},
		},
	})
	return err
}

func (h *ProfileHandler) doAction(b *gotgbot.Bot, ctx *ext.Context) error {
	action := strings.TrimPrefix(ctx.CallbackQuery.Data, "profile_action_")
	data := strings.Split(action, "_")
	switch data[0] {
	case "delete":
		return h.deleteProfile(b, ctx, data[1])
	case "rename":
		return h.renameProfile(b, ctx, data[1])
	case "enable":
		return h.enableProfile(b, ctx, data[1])
	}
	return nil
}

func (h *ProfileHandler) deleteProfile(b *gotgbot.Bot, ctx *ext.Context, ICCID string) error {
	m, err := modem.GetManager().GetModem(h.modemId)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDelete(ICCID); err != nil {
		return err
	}
	_, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile deleted", nil)
	return err
}

func (h *ProfileHandler) enableProfile(b *gotgbot.Bot, ctx *ext.Context, ICCID string) error {
	m, err := modem.GetManager().GetModem(h.modemId)
	if err != nil {
		return err
	}
	m.Lock()
	defer m.Unlock()
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileEnable(ICCID); err != nil {
		return err
	}
	_, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile enabled", nil)
	return err
}

func (h *ProfileHandler) renameProfile(b *gotgbot.Bot, ctx *ext.Context, ICCID string) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, "Enter new name", nil)
	if err != nil {
		return err
	}
	return handlers.NextConversationState(ProfileStateActionRename)
}

func (h *ProfileHandler) handleActionRename(b *gotgbot.Bot, ctx *ext.Context) error {
	fmt.Println("rename")
	return nil
	// m, err := modem.GetManager().GetModem(h.modemId)
	// if err != nil {
	// 	return err
	// }
	// m.Lock()
	// defer m.Unlock()
	// usbDevice, err := m.GetAtPort()
	// if err != nil {
	// 	return err
	// }
	// timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	// if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileSetNickname(h.data[ctx.EffectiveChat.Id], ctx.EffectiveMessage.Text); err != nil {
	// 	return err
	// }
	// _, err = b.SendMessage(ctx.EffectiveChat.Id, "Profile renamed", nil)
	// return err
}

func (h *ProfileHandler) profileMessage(profiles []lpac.Profile) (string, [][]gotgbot.InlineKeyboardButton) {
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
