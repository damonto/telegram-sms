package main

import (
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slog"
)

type HandlerFunc = func(message *tgbotapi.Message, modem Modem) error

type Handler struct {
	chatId int64
	sim    string
	tgbot  *tgbotapi.BotAPI
	modem  Modem

	commands map[string]HandlerFunc
}

func NewHandler(sim string, chatId int64, tgbot *tgbotapi.BotAPI, modem Modem) *Handler {
	return &Handler{
		chatId:   chatId,
		sim:      sim,
		tgbot:    tgbot,
		modem:    modem,
		commands: make(map[string]HandlerFunc, 1),
	}
}

func (h *Handler) RegisterCommand(command string, handler HandlerFunc) {
	h.commands[command] = handler
}

func (h *Handler) Run(command string, message *tgbotapi.Message) error {
	if handler, ok := h.commands[command]; ok {
		return handler(message, h.modem)
	}

	return errors.New("command not found")
}

func (h *Handler) ChatId(message *tgbotapi.Message, modem Modem) error {
	return h.sendText(message.Chat.ID, fmt.Sprintf("%d", message.Chat.ID))
}

func (h *Handler) Sim(message *tgbotapi.Message, modem Modem) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	iccid, _ := modem.GetIccid()
	imei, _ := modem.GetImei()
	signalQuality, _ := modem.GetSignalQuality()

	return h.sendText(message.Chat.ID, h.formatText(fmt.Sprintf("ICCID: %s\nIMEI: %s\nSignal Quality: %d", iccid, imei, signalQuality)))
}

func (h *Handler) RunUSSDCommand(message *tgbotapi.Message, modem Modem) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 2 {
		return errors.New("invalid arguments")
	}

	if err := h.checkSIM(arguments[0]); err != nil {
		return err
	}

	result, err := h.modem.RunUSSDCommand(arguments[1])
	if err != nil {
		return err
	}

	return h.sendText(message.Chat.ID, h.formatText(result))
}

func (h *Handler) SendSms(message *tgbotapi.Message, modem Modem) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 2 {
		return errors.New("invalid arguments")
	}

	if err := h.checkSIM(arguments[0]); err != nil {
		return err
	}

	if err := modem.SendSMS(arguments[1], strings.Join(arguments[2:], " ")); err != nil {
		return h.sendText(message.Chat.ID, h.formatText(err.Error()))
	}

	return nil
}

func (h *Handler) checkSIM(sim string) error {
	if h.sim == sim {
		return nil
	}

	return errors.New("sim does not match")
}

func (h *Handler) checkChatId(chatId int64) error {
	if h.chatId == chatId {
		return nil
	}

	return errors.New("chat id does not match")
}

func (h *Handler) formatText(text string) string {
	carrier, err := h.modem.GetCarrier()
	if err != nil {
		slog.Error("failed to get carrier name", "error", err)
		return text
	}

	return fmt.Sprintf("[%s] %s\n%s", h.sim, carrier, text)
}

func (h *Handler) sendText(chatId int64, message string) error {
	if _, err := h.tgbot.Send(tgbotapi.NewMessage(chatId, message)); err != nil {
		return err
	}

	return nil
}
