package handler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/damonto/libeuicc-go"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"gopkg.in/telebot.v3"
)

type ProfileHandler struct {
	handler
	Iccid string
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
	l, err := h.GetLPA()
	if err != nil {
		return err
	}
	defer l.Close()
	profiles, err := l.GetProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return c.Send("No profiles found.")
	}

	message, buttons := h.toTextMessage(c, profiles)
	return c.Send(message, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdownV2,
		ReplyMarkup: buttons,
	})
}

func (h *ProfileHandler) toTextMessage(c telebot.Context, profiles []*libeuicc.Profile) (string, *telebot.ReplyMarkup) {
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
		if p.State == libeuicc.ProfileStateEnabled {
			emoji = "✅"
		} else {
			emoji = "🅾️"
		}
		message += fmt.Sprintf(template, emoji, util.EscapeText(name), p.Iccid)
		btn := selector.Data(fmt.Sprintf("%s (%s)", name, p.Iccid[len(p.Iccid)-4:]), fmt.Sprint(time.Now().UnixNano()), p.Iccid)
		c.Bot().Handle(&btn, func(c telebot.Context) error {
			h.Iccid = c.Data()
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
		return c.Send("Invalid action.")
	}
}

func (h *ProfileHandler) handleAskAction(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()

	l, err := h.GetLPA()
	if err != nil {
		return err
	}
	defer l.Close()
	profile, err := l.FindProfile(h.Iccid)
	if err != nil {
		return err
	}

	buttons := []telebot.ReplyButton{
		{
			Text: ProfileActionRename,
		},
	}
	if profile.State == libeuicc.ProfileStateDisabled {
		buttons = append(buttons, telebot.ReplyButton{
			Text: ProfileActionEnable,
		}, telebot.ReplyButton{
			Text: ProfileActionDelete,
		})
	}

	template := `
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
	if profile.State == libeuicc.ProfileStateEnabled {
		emoji = "✅"
	} else {
		emoji = "🅾️"
	}
	return c.Send(fmt.Sprintf(template, emoji, util.EscapeText(name), fmt.Sprintf("`%s`", profile.Iccid)), &telebot.SendOptions{
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
	l, err := h.GetLPA()
	if err != nil {
		return err
	}
	defer l.Close()
	if err := l.Delete(h.Iccid); err != nil {
		return err
	}
	return c.Send("Your profile has been deleted. /profiles")
}

func (h *ProfileHandler) handleActionEnable(c telebot.Context) error {
	h.modem.Lock()
	l, err := h.GetLPA()
	if err != nil {
		return err
	}
	if err := l.EnableProfile(h.Iccid, false); err != nil {
		h.modem.Unlock()
		return err
	}
	l.Close()
	h.modem.Unlock()
	// Sometimes the modem needs to be restarted to apply the changes.
	if err := h.modem.Restart(); err != nil {
		slog.Error("unable to restart modem, you may need to restart this modem manually.", "error", err)
	}
	return c.Send("Your profile has been enabled. Please wait a moment for it to take effect. /profiles")
}

func (h *ProfileHandler) handleActionRename(c telebot.Context) error {
	h.modem.Lock()
	defer h.modem.Unlock()

	l, err := h.GetLPA()
	if err != nil {
		return err
	}
	defer l.Close()
	if err := l.SetNickname(h.Iccid, c.Text()); err != nil {
		return err
	}
	return c.Send(fmt.Sprintf("Your profile has been renamed to *%s*\\. /profiles", util.EscapeText(c.Text())), &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdownV2,
	})
}
