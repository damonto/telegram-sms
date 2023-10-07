package main

import (
	"flag"
	"fmt"
	"os/user"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/exp/slog"
)

var (
	tgBotToken          string
	tgChatId            int64
	enableEuiccFeatures bool
	enableDebugMode     bool
)

func init() {
	flag.StringVar(&tgBotToken, "token", "", "Telegram API token")
	flag.Int64Var(&tgChatId, "chat-id", 0, "Your telegram chat id")
	flag.BoolVar(&enableEuiccFeatures, "euicc", false, "Enable eUICC features")
	flag.BoolVar(&enableDebugMode, "debug", false, "Show verbose info")
	flag.Parse()
}

func main() {
	user, err := user.Current()
	if err != nil {
		slog.Error("unable to find current running user", "error", err)
		panic(err)
	}

	if user.Username != "root" {
		slog.Error("must be running as the root user")
		return
	}

	bot, err := tgbotapi.NewBotAPI(tgBotToken)
	if err != nil {
		slog.Error("failed to connect to telegram bot", "error", err)
		panic(err)
	}
	bot.Debug = enableDebugMode

	modem, err := NewModem()
	if err != nil {
		slog.Error("failed to connect to modem", "error", err)
		panic(err)
	}

	handler := NewHandler(tgChatId, enableEuiccFeatures, bot, modem)
	handler.RegisterCommands()

	go func(bot *tgbotapi.BotAPI, modem Modem) {
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 30
		updates := bot.GetUpdatesChan(updateConfig)

		for update := range updates {
			if update.Message != nil && update.Message.IsCommand() {
				slog.Info("command received", "command", update.Message.Command(), "raw", update.Message.Text)
				if err := handler.HandleCommand(update.Message.Command(), update.Message); err != nil {
					slog.Error("failed to handle command", "error", err)
					msg := tgbotapi.NewMessage(tgChatId, err.Error())
					msg.ReplyToMessageID = update.Message.MessageID
					if _, err := bot.Send(msg); err != nil {
						slog.Error("failed to send message", "text", msg.Text, "error", err)
					}
				}
				continue
			}

			if update.CallbackQuery != nil {
				slog.Info("command callback received", "button", update.CallbackQuery.Data)
				if err := handler.HandleCallback(update.CallbackQuery); err != nil {
					slog.Error("failed to handle callback", "error", err)
					msg := tgbotapi.NewMessage(tgChatId, err.Error())
					msg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
					if _, err := bot.Send(msg); err != nil {
						slog.Error("failed to send message", "text", msg.Text, "error", err)
					}
				}
				continue
			}

			if err := handler.HandleRawMessage(update.Message); err != nil {
				slog.Error("failed to handle callback", "error", err)
			}
		}
	}(bot, modem)

	modem.SubscribeSMS(func(modem modemmanager.Modem, sms modemmanager.Sms) {
		sender, err := sms.GetNumber()
		if err != nil {
			slog.Error("failed to get phone number", "error", err)
			return
		}

		three3gpp, _ := modem.Get3gpp()
		operatorName, err := three3gpp.GetOperatorName()
		if err != nil {
			slog.Error("failed to get operator name", "error", err)
			return
		}

		text, err := sms.GetText()
		if err != nil {
			slog.Error("failed to get SMS text", "error", err)
			return
		}

		slog.Info("SMS received", "operatorName", operatorName, "sender", sender, "text", text)
		msg := tgbotapi.NewMessage(tgChatId, fmt.Sprintf("*\\[%s\\] %s*\n%s", operatorName, EscapeText(sender), EscapeText(text)))
		msg.ParseMode = "markdownV2"
		if _, err := bot.Send(msg); err != nil {
			slog.Error("failed to send message", "text", msg.Text, "error", err)
		}
	})
}
