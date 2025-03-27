package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/damonto/telegram-sms/internal/app"
	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/godbus/dbus/v5"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var Version string

type Subscriber struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func init() {
	flag.StringVar(&config.C.BotToken, "bot-token", "", "Telegram bot token")
	flag.Var(&config.C.AdminId, "admin-id", "Admin user ID with bot management privileges")
	flag.BoolVar(&config.C.Slowdown, "slowdown", false, "Enable slowdown mode (MSS: 120)")
	flag.BoolVar(&config.C.Compatible, "compatible", false, "Enable if your modem does not support proactive refresh")
	flag.StringVar(&config.C.Endpoint, "endpoint", "https://api.telegram.org", "Telegram Bot API endpoint")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()
}

func main() {
	if config.C.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if os.Geteuid() != 0 {
		slog.Error("Please run as root")
		os.Exit(1)
	}
	if err := config.C.IsValid(); err != nil {
		slog.Error("Config is invalid", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting telegram SMS bot", "version", Version)

	bot, err := telego.NewBot(config.C.BotToken,
		telego.WithAPIServer(config.C.Endpoint),
		telego.WithDefaultLogger(config.C.Verbose, true),
	)
	if err != nil {
		panic(err)
	}
	mm, err := modem.NewManager()
	if err != nil {
		panic(err)
	}
	go subscribe(bot, mm)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	me, err := bot.GetMe(ctx)
	if err != nil {
		panic(err)
	}
	slog.Info("Bot started", "username", me.Username, "id", me.ID)

	app, err := app.New(ctx, bot, mm)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := app.Start(); err != nil {
			panic(err)
		}
	}()
	<-ctx.Done()
	slog.Info("Stopping telegram SMS bot")
	app.Shutdown()
	slog.Info("Goodbye!")
}

func subscribe(bot *telego.Bot, mm *modem.Manager) {
	var err error
	subscribers := make(map[dbus.ObjectPath]*Subscriber)
	modems, err := mm.Modems()
	if err != nil {
		panic(err)
	}

	go subscribeMessaging(bot, modems, subscribers)

	err = mm.Subscribe(func(modems map[dbus.ObjectPath]*modem.Modem) error {
		for path, s := range subscribers {
			slog.Debug("Canceling subscriber", "path", path)
			s.cancel()
		}
		go subscribeMessaging(bot, modems, subscribers)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func subscribeMessaging(bot *telego.Bot, modems map[dbus.ObjectPath]*modem.Modem, subscribers map[dbus.ObjectPath]*Subscriber) {
	for path, m := range modems {
		slog.Info("Subscribing to modem messaging", "path", path)
		ctx, cancel := context.WithCancel(context.Background())
		go func(ctx context.Context, m *modem.Modem) {
			if err := m.SubscribeMessaging(ctx, func(message *modem.SMS) error {
				if err := send(bot, m, message); err != nil {
					slog.Error("Failed to send message", "error", err)
				}
				return nil
			}); err != nil {
				slog.Error("Failed to subscribe to modem messaging", "error", err)
			}
		}(ctx, m)
		subscribers[path] = &Subscriber{ctx: ctx, cancel: cancel}
	}
}

func send(bot *telego.Bot, modem *modem.Modem, messsage *modem.SMS) error {
	template := `
[ ] *\[%s\] \- %s*
> %s
`
	operatorName, err := modem.OperatorName()
	if err != nil {
		slog.Error("Failed to get operator name", "error", err)
		operatorName = "unknown"
	}
	message := fmt.Sprintf(
		template,
		util.EscapeText(operatorName),
		util.EscapeText(messsage.Number),
		fmt.Sprintf("`%s`", util.EscapeText(messsage.Text)),
	)
	for _, adminId := range config.C.AdminId.MarshalInt64() {
		msg, err := bot.SendMessage(context.Background(), tu.Message(
			tu.ID(adminId),
			message,
		).WithParseMode(telego.ModeMarkdownV2))
		if err != nil {
			slog.Error("Failed to send message", "error", err, "to", adminId, "message", message)
		}
		slog.Info("Message sent", "id", msg.Chat.ID, "to", adminId)
	}
	return nil
}
