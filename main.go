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
	token  string
	sim    string
	chatId int64
	debug  bool
)

func init() {
	flag.StringVar(&token, "token", "", "Telegram API token")
	flag.StringVar(&sim, "sim", "", "A unqiue Id of a SIM")
	flag.Int64Var(&chatId, "chat-id", 0, "Your telegram chat id")
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
		slog.Error("telegram messenger must be running as the root user")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		slog.Error("failed to connect to telegram bot", "error", err)
		panic(err)
	}
	bot.Debug = debug

	modem, err := NewModem()
	if err != nil {
		slog.Error("failed to connect to modem", "error", err)
		panic(err)
	}

	go func(bot *tgbotapi.BotAPI, modem Modem) {
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 30
		updates := bot.GetUpdatesChan(updateConfig)
		handler := NewHandler(sim, chatId, bot, modem)

		for update := range updates {
			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}

			handler.RegisterCommand("sim", handler.Sim)
			handler.RegisterCommand("chatid", handler.ChatId)
			handler.RegisterCommand("send", handler.SendSms)
			handler.RegisterCommand("ussd", handler.RunUSSDCommand)

			slog.Info("command received", "command", update.Message.Command(), "raw", update.Message.Text)
			if err := handler.Run(update.Message.Command(), update.Message); err != nil {
				slog.Error("failed to run command", "error", err)
				continue
			}
		}
	}(bot, modem)

	modem.SMSRecevied(func(sms modemmanager.Sms) {
		form, err := sms.GetNumber()
		if err != nil {
			slog.Error("failed to get phone number", "error", err)
			return
		}

		text, err := sms.GetText()
		if err != nil {
			slog.Error("failed to get SMS text", "error", err)
			return
		}

		msg := tgbotapi.NewMessage(chatId, formatText(fmt.Sprintf("From: %s\n%s", form, text), modem))
		if _, err := bot.Send(msg); err != nil {
			slog.Error("failed to send message", "text", msg.Text)
		}
	})
}

func formatText(text string, modem Modem) string {
	carrier, err := modem.GetCarrier()
	if err != nil {
		slog.Error("failed to get carrier name", "error", err)
		return text
	}

	return fmt.Sprintf("[%s] %s\n%s", sim, carrier, text)
}
