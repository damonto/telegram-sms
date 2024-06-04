package middleware

import (
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/damonto/telegram-sms/config"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
)

func RequiredAdmin(bot *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveUser.Id == config.C.AdminId {
		return nil
	}
	return ErrPermissionDenied
}
