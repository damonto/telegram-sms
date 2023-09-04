package main

import (
	"flag"
	"fmt"
	"os/user"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/exp/slog"
)

var (
	token      string
	chatId     int64
	modemIndex int
	debug      bool
)

func init() {
	flag.StringVar(&token, "token", "", "Telegram API token")
	flag.Int64Var(&chatId, "chat-id", 0, "Your telegram chat id")
	flag.IntVar(&modemIndex, "modem", 0, "The modem index number")
	flag.BoolVar(&debug, "debug", false, "Show verbose info")
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

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		slog.Error("failed to connect to telegram bot", "error", err)
		panic(err)
	}
	bot.Debug = debug

	modem, err := NewModem(modemIndex)
	if err != nil {
		slog.Error("failed to connect to modem", "error", err)
		panic(err)
	}

	handler := NewHandler(chatId, bot, modem)
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
					msg := tgbotapi.NewMessage(chatId, escapeText(err.Error()))
					if _, err := bot.Send(msg); err != nil {
						slog.Error("failed to send message", "text", msg.Text, "error", err)
					}
					continue
				}
			}

			if update.CallbackQuery != nil {
				slog.Info("command callback received", "button", update.CallbackQuery.Data)
				if err := handler.HandleCallback(update.CallbackQuery); err != nil {
					slog.Error("failed to handle callback", "error", err)
					msg := tgbotapi.NewMessage(chatId, escapeText(err.Error()))
					if _, err := bot.Send(msg); err != nil {
						slog.Error("failed to send message", "text", msg.Text, "error", err)
					}
					continue
				}
			}
		}
	}(bot, modem)

	modem.SubscribeSMS(func(sms modemmanager.Sms) {
		sender, err := sms.GetNumber()
		if err != nil {
			slog.Error("failed to get phone number", "error", err)
			return
		}

		operator, err := modem.GetOperatorName()
		if err != nil {
			slog.Error("failed to get operator name", "error", err)
			return
		}

		text, err := sms.GetText()
		if err != nil {
			slog.Error("failed to get SMS text", "error", err)
			return
		}

		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("*\\[%s\\] %s*\n%s", operator, escapeText(sender), escapeText(text)))
		msg.ParseMode = "markdownV2"
		if _, err := bot.Send(msg); err != nil {
			slog.Error("failed to send message", "text", msg.Text, "error", err)
		}
	})
}

func escapeText(text string) string {
	replacements := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement, "\\"+replacement)
	}
	return text
}
