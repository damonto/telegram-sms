package handler

import (
	"fmt"
	"log/slog"

	sgp22 "github.com/damonto/euicc-go/v2"
	"github.com/damonto/telegram-sms/internal/app/state"
	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type ProfileHandler struct {
	*Handler
}

const (
	ProfileActionCallbackDataPrefix = "profile"

	ProfileMessageTemplate = `
%s *%s*
%s
	`

	ProfileActionSetNickname state.State = "Set Nickname"
	ProfileActionDelete      state.State = "Delete"
	ProfileActionEnable      state.State = "Enable"
)

type ProfileValue struct {
	ICCID   string
	Action  state.State
	Profile *sgp22.ProfileInfo
	Value   string
	Modem   *modem.Modem
}

func NewProfileHandler() state.Handler {
	h := new(ProfileHandler)
	return h
}

func (h *ProfileHandler) HandleCallbackQuery(ctx *th.Context, query telego.CallbackQuery, state *state.ChatState) error {
	var err error
	value := state.Value.(*ProfileValue)
	value.ICCID = query.Data[len(ProfileActionCallbackDataPrefix)+1:]
	l, err := lpa.New(value.Modem)
	if err != nil {
		return err
	}
	id, _ := sgp22.NewICCID(value.ICCID)
	ps, err := l.ListProfile(id)
	if err != nil {
		return err
	}
	defer l.Close()
	value.Profile = ps[0]
	return h.sendActionMessage(ctx, query, ps[0])
}

func (h *ProfileHandler) sendActionMessage(ctx *th.Context, query telego.CallbackQuery, profile *sgp22.ProfileInfo) error {
	var buttons []telego.KeyboardButton
	if profile.ProfileState == sgp22.ProfileEnabled {
		buttons = tu.KeyboardRow(tu.KeyboardButton(string(ProfileActionSetNickname)))
	} else {
		buttons = tu.KeyboardRow(
			tu.KeyboardButton(string(ProfileActionSetNickname)),
			tu.KeyboardButton(string(ProfileActionEnable)),
			tu.KeyboardButton(string(ProfileActionDelete)),
		)
	}

	var message string
	name := fmt.Sprintf("[%s] %s",
		profile.ServiceProviderName,
		util.If(profile.ProfileNickname != "", profile.ProfileNickname, profile.ProfileName),
	)
	message += fmt.Sprintf(ProfileMessageTemplate,
		util.If(profile.ProfileState == sgp22.ProfileEnabled, "✅", "🅾️"),
		util.EscapeText(name),
		profile.ICCID,
	)
	message = util.EscapeText("What do you want to do with the profile? \n" + message)
	_, err := h.ReplyCallbackQuery(ctx, query, message, func(message *telego.SendMessageParams) error {
		message.WithReplyMarkup(tu.Keyboard(buttons))
		return nil
	})
	return err
}

func (h *ProfileHandler) HandleMessage(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	if state.State(message.Text) == ProfileActionSetNickname {
		return h.askNickname(ctx, message, s)
	}
	if s.State == ProfileActionSetNickname {
		return h.setNickname(ctx, message, s)
	}
	if state.State(message.Text) == ProfileActionEnable {
		return h.enableProfile(ctx, message, s)
	}
	if state.State(message.Text) == ProfileActionDelete {
		return h.confirmDelete(ctx, message, s)
	}
	if s.State == ProfileActionDelete {
		return h.deleteProfile(ctx, message, s)
	}
	state.M.Exit(message.Chat.ID)
	return nil
}

func (h *ProfileHandler) deleteProfile(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	if message.Text != "Yes" {
		_, err := h.ReplyMessage(ctx, message, util.EscapeText("Okay, the profile will not be deleted."), nil)
		return err
	}
	value := s.Value.(*ProfileValue)
	l, err := lpa.New(value.Modem)
	if err != nil {
		return err
	}
	defer l.Close()
	id, _ := sgp22.NewICCID(value.ICCID)
	if err := l.Delete(id); err != nil {
		return err
	}
	_, err = h.ReplyMessage(ctx, message, util.EscapeText("The profile has been deleted. /profiles"), nil)
	return err
}

func (h *ProfileHandler) confirmDelete(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	state.M.Current(message.Chat.ID, ProfileActionDelete)
	value := s.Value.(*ProfileValue)
	_, err := h.ReplyMessage(
		ctx,
		message,
		util.EscapeText(
			fmt.Sprintf(
				"Are you sure you want to delete the profile %s?",
				util.If(value.Profile.ProfileNickname != "", value.Profile.ProfileNickname, value.Profile.ProfileName),
			),
		),
		func(m *telego.SendMessageParams) error {
			m.WithReplyMarkup(tu.Keyboard(
				tu.KeyboardRow(
					tu.KeyboardButton("Yes"),
					tu.KeyboardButton("No"),
				),
			))
			return nil
		},
	)
	return err
}

func (h *ProfileHandler) enableProfile(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	value := s.Value.(*ProfileValue)
	l, err := lpa.New(value.Modem)
	if err != nil {
		return err
	}
	defer l.Close()
	id, _ := sgp22.NewICCID(value.ICCID)
	if err := l.EnableProfile(id); err != nil {
		return err
	}
	if err := value.Modem.Restart(); err != nil {
		slog.Warn("Failed to restart the modem", "error", err)
	}
	_, err = h.ReplyMessage(
		ctx,
		message,
		util.EscapeText("The profile has been enabled. It may take a few seconds for the profile to be activated."),
		nil,
	)
	return err
}

func (h *ProfileHandler) setNickname(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	value := s.Value.(*ProfileValue)
	value.Value = message.Text
	l, err := lpa.New(value.Modem)
	if err != nil {
		return err
	}
	defer l.Close()
	id, _ := sgp22.NewICCID(value.ICCID)
	if err := l.SetNickname(id, value.Value); err != nil {
		return err
	}
	_, err = h.ReplyMessage(
		ctx,
		message,
		util.EscapeText("The nickname has been updated, you can now use the /profiles command to see the changes."),
		nil,
	)
	return err
}

func (h *ProfileHandler) askNickname(ctx *th.Context, message telego.Message, s *state.ChatState) error {
	state.M.Current(message.Chat.ID, ProfileActionSetNickname)
	_, err := h.ReplyMessage(
		ctx,
		message,
		util.EscapeText("Okay, please enter the new nickname for the profile."),
		nil,
	)
	return err
}

func (h *ProfileHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		l, err := h.LPA(ctx)
		if err != nil {
			return err
		}
		defer l.Close()

		state.M.Enter(update.Message.Chat.ID, &state.ChatState{
			Handler: h,
			Value: &ProfileValue{
				Modem: h.Modem(ctx),
			},
		})

		profiles, err := l.ListProfile(nil)
		if err != nil {
			return err
		}
		buttons, message := h.message(profiles)
		_, err = h.Reply(ctx, update, message, func(message *telego.SendMessageParams) error {
			message.WithReplyMarkup(buttons)
			return nil
		})
		return err
	}
}

func (h *ProfileHandler) message(profiles []*sgp22.ProfileInfo) (*telego.InlineKeyboardMarkup, string) {
	var message string
	var buttons [][]telego.InlineKeyboardButton
	for _, profile := range profiles {
		name := fmt.Sprintf("[%s] %s",
			profile.ServiceProviderName,
			util.If(profile.ProfileNickname != "", profile.ProfileNickname, profile.ProfileName),
		)
		message += fmt.Sprintf(ProfileMessageTemplate,
			util.If(profile.ProfileState == sgp22.ProfileEnabled, "✅", "🅾️"),
			util.EscapeText(name),
			profile.ICCID,
		)
		buttons = append(buttons, tu.InlineKeyboardRow(telego.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s (%s)", name, profile.ICCID[len(profile.ICCID)-3:]),
			CallbackData: fmt.Sprintf("%s:%s", ProfileActionCallbackDataPrefix, profile.ICCID),
		}))
	}
	return tu.InlineKeyboard(buttons...), message
}
