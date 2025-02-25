package app

import (
	"context"
	"time"

	"github.com/damonto/telegram-sms/internal/app/router"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type application struct {
	Bot     *telego.Bot
	handler *th.BotHandler
	updates <-chan telego.Update
	ctx     context.Context
}

func NewApp(bot *telego.Bot, ctx context.Context) (*application, error) {
	app := &application{
		ctx: ctx,
	}
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
	app.registerMiddleware()
	app.registerRouter()
	return app.handler.Start()
}

func (app *application) registerRouter() {
	router.NewRouter(app.handler).Register()
}

func (app *application) registerMiddleware() {
	app.handler.Use(th.PanicRecovery())
}

func (app *application) Shutdown() {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second*30)
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
