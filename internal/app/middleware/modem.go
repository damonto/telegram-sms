package middleware

import (
	"errors"
	"fmt"

	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

var (
	ErrNextHandlerNotSet = errors.New("next handler not set")
	ErrNoEuiccModemFound = errors.New("no eUICC modem found")
)

func SelectModem(requiredEuicc bool) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			modems, err := modems(requiredEuicc)
			if err != nil {
				return err
			}
			if len(modems) == 1 {
				for _, m := range modems {
					c.Set("modem", m)
					return next(c)
				}
			}

			done := make(chan string, 1)
			defer close(done)
			if err := selectModem(c, modems, done); err != nil {
				return err
			}
			c.Set("modem", modems[<-done])
			return next(c)
		}
	}
}

func selectModem(c telebot.Context, modems map[string]*modem.Modem, done chan string) error {
	selector := telebot.ReplyMarkup{}
	btns := make([]telebot.Btn, 0, len(modems))
	for k, m := range modems {
		model, _ := m.GetModel()
		btn := selector.Data(fmt.Sprintf("%s (%s)", model, k), uuid.New().String(), k)
		c.Bot().Handle(&btn, func(c telebot.Context) error {
			done <- c.Callback().Data
			return c.Delete()
		})
		btns = append(btns, btn)
	}
	selector.Inline(selector.Row(btns...))
	return c.Send("I found the following modems, please select one:", &selector)
}

func modems(requiredEuicc bool) (map[string]*modem.Modem, error) {
	modems := modem.GetManager().GetModems()
	if len(modems) == 0 {
		return nil, modem.ErrModemNotFound
	}

	if requiredEuicc {
		for k, m := range modems {
			if !m.IsEuicc {
				delete(modems, k)
			}
		}
		if len(modems) == 0 {
			return nil, ErrNoEuiccModemFound
		}
	}
	return modems, nil
}
