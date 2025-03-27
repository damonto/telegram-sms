package handler

import (
	"fmt"
	"strconv"
	"strings"

	sgp22 "github.com/damonto/euicc-go/v2"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type SendNotificationHandler struct {
	*Handler
}

func NewSendNotificationHandler() *SendNotificationHandler {
	h := new(SendNotificationHandler)
	return h
}

func (h *SendNotificationHandler) Handle() th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		lpa, err := h.LPA(ctx)
		if err != nil {
			return err
		}
		defer lpa.Close()
		if update.Message.Text == "" {
			_, err := h.Reply(ctx, update, util.EscapeText("Please provide a sequence number."), nil)
			return err
		}
		seq, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/send_notification "))
		if err != nil {
			_, err := h.Reply(ctx, update, util.EscapeText("Invalid sequence number."), nil)
			return err
		}
		if err := lpa.SendNotification(sgp22.SequenceNumber(seq)); err != nil {
			_, err := h.Reply(ctx, update, fmt.Sprintf("Failed to send notification %s", util.EscapeText(err.Error())), nil)
			return err
		}
		_, err = h.Reply(ctx, update, util.EscapeText(fmt.Sprintf("The notification %d has been sent.", seq)), nil)
		return err
	}
}
