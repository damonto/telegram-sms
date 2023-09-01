package main

import (
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc = func(message *tgbotapi.Message) error

type Handler struct {
	chatId int64
	tgbot  *tgbotapi.BotAPI
	modem  Modem

	commands map[string]HandlerFunc
}

func NewHandler(chatId int64, tgbot *tgbotapi.BotAPI, modem Modem) *Handler {
	return &Handler{
		chatId:   chatId,
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
		return handler(message)
	}

	return errors.New("command not found")
}

func (h *Handler) ChatId(message *tgbotapi.Message) error {
	return h.sendText(message.Chat.ID, fmt.Sprintf("%d", message.Chat.ID))
}

func (h *Handler) Sim(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	iccid, _ := h.modem.GetIccid()
	imei, _ := h.modem.GetImei()
	signalQuality, _ := h.modem.GetSignalQuality()

	return h.sendText(message.Chat.ID, fmt.Sprintf("ICCID: %s\nIMEI: %s\nSignal Quality: %d", iccid, imei, signalQuality))
}

func (h *Handler) RunUSSDCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 1 {
		return errors.New("invalid arguments")
	}

	result, err := h.modem.RunUSSDCommand(arguments[0])
	if err != nil {
		return err
	}

	return h.sendText(message.Chat.ID, result)
}

func (h *Handler) SendSms(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 2 {
		return errors.New("invalid arguments")
	}

	if err := h.modem.SendSMS(arguments[0], strings.Join(arguments[1:], " ")); err != nil {
		return h.sendText(message.Chat.ID, err.Error())
	}

	return nil
}

func (h *Handler) checkChatId(chatId int64) error {
	if h.chatId == chatId {
		return nil
	}

	return errors.New("chat id does not match")
}

func (h *Handler) sendText(chatId int64, message string) error {
	if _, err := h.tgbot.Send(tgbotapi.NewMessage(chatId, message)); err != nil {
		return err
	}

	return nil
}
