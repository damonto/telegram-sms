package middleware

import (
	"errors"

	"slices"

	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

var ErrPermissionDenied = errors.New("permission denied")

func Admin() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		if !slices.Contains(config.C.AdminId.UnmarshalInt64(), update.Message.From.ID) {
			return ErrPermissionDenied
		}
		return ctx.Next(update)
	}
}
