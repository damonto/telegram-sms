package app

import (
	"context"
	"time"

	"github.com/damonto/telegram-sms/internal/app/router"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type application struct {
	Bot     *telego.Bot
	m       *modem.Manager
	handler *th.BotHandler
	updates <-chan telego.Update
	ctx     context.Context
}

func NewApp(ctx context.Context, bot *telego.Bot, m *modem.Manager) (*application, error) {
	app := &application{Bot: bot, m: m, ctx: ctx}
	var err error
	app.updates, err = bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return nil, err
	}
	app.handler, err = th.NewBotHandler(bot, app.updates)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (app *application) Start() error {
	app.handler.Use(th.PanicRecovery())
	router.NewRouter(app.Bot, app.handler, app.m).Register()
	return app.handler.Start()
}

func (app *application) Shutdown() {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer stopCancel()

outer:
	for len(app.updates) > 0 {
		select {
		case <-stopCtx.Done():
			break outer
		case <-time.After(100 * time.Microsecond):
			//
		}
	}
	app.handler.StopWithContext(stopCtx)
}
