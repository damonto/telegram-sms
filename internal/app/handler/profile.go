package handler

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func Profile() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		return nil
	}
}
