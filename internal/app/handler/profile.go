package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type ProfileHandler struct {
	handler
	ICCID string
}

const (
	StateProfileHandleAction = "profile_handle_action"
	StateProfileActionDelete = "profile_delete"
	StateProfileActionRename = "profile_rename"

	ProfileActionRename = "Rename"
	ProfileActionDelete = "Delete"
	ProfileActionEnable = "Enable"
)

func HandleProfilesCommand(c telebot.Context) error {
	h := &ProfileHandler{}
	h.init(c)
	h.state = h.stateManager.New(c)
	h.state.States(map[string]telebot.HandlerFunc{
		StateProfileHandleAction: h.handleAction,
		StateProfileActionRename: h.handleActionRename,
		StateProfileActionDelete: h.handleActionDelete,
	})
	return h.handle(c)
}

func (h *ProfileHandler) handle(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()
	usbDevice, err := h.modem.GetAtPort()
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
		return c.Send("No profiles found.")
	}

	message, buttons := h.toTextMessage(c, profiles)
	return c.Send(util.EscapeText(message), &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: buttons,
	})
}

func (h *ProfileHandler) toTextMessage(c telebot.Context, profiles []*lpac.Profile) (string, *telebot.ReplyMarkup) {
	selector := &telebot.ReplyMarkup{}
	template := `
%s *%s*
%s
	`
	var message string
	buttons := make([]telebot.Btn, 0, len(profiles))
	for _, p := range profiles {
		name := fmt.Sprintf("[%s] ", p.ProviderName)
		if p.Nickname != "" {
			name += p.Nickname
		} else {
			name += p.ProfileName
		}
		var emoji string
		if p.State == lpac.ProfileStateEnabled {
			emoji = "✅"
		} else {
			emoji = "🅾️"
		}
		message += fmt.Sprintf(template, emoji, name, p.ICCID)
		btn := selector.Data(fmt.Sprintf("%s (%s)", name, p.ICCID[len(p.ICCID)-4:]), fmt.Sprint(time.Now().UnixNano()), p.ICCID)
		c.Bot().Handle(&btn, func(c telebot.Context) error {
			h.ICCID = c.Data()
			h.state.Next(StateProfileHandleAction)
			return h.handleAskAction(c)
		})
		buttons = append(buttons, btn)
	}
	selector.Inline(selector.Split(1, buttons)...)
	return message, selector
}

func (h *ProfileHandler) handleAction(c telebot.Context) error {
	switch c.Text() {
	case ProfileActionEnable:
		return h.handleActionEnable(c)
	case ProfileActionRename:
		h.state.Next(StateProfileActionRename)
		return c.Send("OK. Send me the new name.")
	case ProfileActionDelete:
		h.state.Next(StateProfileActionDelete)
		return c.Send("Are you sure you want to delete this profile?", &telebot.ReplyMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			ReplyKeyboard: [][]telebot.ReplyButton{
				{
					{
						Text: "Yes",
					},
					{
						Text: "No",
					},
				},
			},
		})
	default:
		return c.Send("Invalid action")
	}
}

func (h *ProfileHandler) handleAskAction(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()
	usbDevice, err := h.modem.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	profile, err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileInfo(h.ICCID)
	if err != nil {
		return err
	}

	buttons := []telebot.ReplyButton{
		{
			Text: ProfileActionRename,
		},
	}
	if profile.State == lpac.ProfileStateDisabled {
		buttons = append(buttons, telebot.ReplyButton{
			Text: ProfileActionEnable,
		}, telebot.ReplyButton{
			Text: ProfileActionDelete,
		})
	}

	template := `
You've selected the profile:
%s *%s*
%s
What do you want to do with this profile?
	`
	name := fmt.Sprintf("[%s] ", profile.ProviderName)
	if profile.Nickname != "" {
		name += profile.Nickname
	} else {
		name += profile.ProfileName
	}
	var emoji string
	if profile.State == lpac.ProfileStateEnabled {
		emoji = "✅"
	} else {
		emoji = "🅾️"
	}
	return c.Send(util.EscapeText(fmt.Sprintf(template, emoji, name, fmt.Sprintf("`%s`", profile.ICCID))), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
		ReplyMarkup: &telebot.ReplyMarkup{
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
			ReplyKeyboard:   [][]telebot.ReplyButton{buttons},
		},
	})
}

func (h *ProfileHandler) handleActionDelete(c telebot.Context) error {
	if c.Text() != "Yes" {
		return c.Send("Canceled! Your profile won't be deleted. /profiles")
	}
	h.modem.Lock()
	defer h.modem.Unlock()
	usbDevice, err := h.modem.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileDelete(h.ICCID); err != nil {
		// FIXME: On some modems, the profile deletion command will return an error even if the profile is deleted.
		if err.Error() != "internal error, maybe illegal iccid/aid coding" {
			return err
		}
	}
	return c.Send("Your profile has been deleted. /profiles")
}

func (h *ProfileHandler) handleActionEnable(c telebot.Context) error {
	h.modem.Lock()
	usbDevice, err := h.modem.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileEnable(h.ICCID); err != nil {
		return err
	}
	h.modem.Unlock()

	// Sometimes the modem needs to be restarted to apply the changes.
	if err := h.modem.Restart(); err != nil {
		slog.Error("unable to restart modem, you may need to restart this modem manually", "error", err)
	}
	return c.Send("Your profile has been enabled. Please wait a moment for it to take effect. /profiles")
}

func (h *ProfileHandler) handleActionRename(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()
	usbDevice, err := h.modem.GetAtPort()
	if err != nil {
		return err
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := lpac.NewCmd(timeoutCtx, usbDevice).ProfileSetNickname(h.ICCID, c.Text()); err != nil {
		return err
	}
	return c.Send(util.EscapeText(fmt.Sprintf("Your profile has been renamed to *%s* . /profiles", c.Text())), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}
