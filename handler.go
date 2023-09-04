package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slog"
)

type HandlerFunc = func(message *tgbotapi.Message) error
type CallbackFunc = func(callback *tgbotapi.CallbackQuery, value string) error

type Handler interface {
	RegisterCommands() error
	HandleCommand(command string, message *tgbotapi.Message) error
	HandleCallback(callback *tgbotapi.CallbackQuery) error
}

type handler struct {
	chatId int64
	tgbot  *tgbotapi.BotAPI
	modem  Modem

	commands map[string]command
}

type command struct {
	command     string
	description string
	handler     HandlerFunc
	callback    CallbackFunc
}

func NewHandler(chatId int64, tgbot *tgbotapi.BotAPI, modem Modem) Handler {
	return &handler{
		chatId: chatId,
		tgbot:  tgbot,
		modem:  modem,
	}
}

func (h *handler) RegisterCommands() error {
	h.commands = map[string]command{
		"chatid":        {command: "chatid", description: "Obtain your chat id", handler: h.handleChatIdCommand},
		"sim":           {command: "sim", description: "Obtain SIM card properties", handler: h.handleSimCommand},
		"switchsimslot": {command: "switchsimslot", description: "Switch to another SIM slot", handler: h.handleSwitchSlotCommand, callback: h.handleSwitchSlotCallback},
		"sms":           {command: "sms", description: "Send an SMS to a phone number", handler: h.handleSendSmsCommand},
		"ussd":          {command: "ussd", description: "Send a USSD command to your SIM card", handler: h.handleUSSDCommand},
		"ussdresponed":  {command: "ussdresponed", description: "Respond the last ussd command", handler: h.handleUSSDRespondCommand},
	}
	botCommands := []tgbotapi.BotCommand{}
	for _, c := range h.commands {
		botCommands = append(botCommands, tgbotapi.BotCommand{
			Command:     c.command,
			Description: c.description,
		})
	}

	response, err := h.tgbot.Request(tgbotapi.NewSetMyCommands(botCommands...))
	if err != nil {
		return err
	}

	slog.Info("set telegram bot commands", "ok", response.Ok, "response", response.Description)
	if !response.Ok {
		return errors.New(response.Description)
	}
	return nil
}

func (h *handler) HandleCommand(command string, message *tgbotapi.Message) error {
	if command == "start" {
		return h.handleStartCommand(message)
	}

	if command, ok := h.commands[command]; ok {
		return command.handler(message)
	}
	return errors.New("command not found")
}

func (h *handler) HandleCallback(callback *tgbotapi.CallbackQuery) error {
	button := strings.Split(callback.Data, ":")

	if command, ok := h.commands[button[0]]; ok {
		return command.callback(callback, button[1])
	}
	return errors.New("command not found")
}

func (h *handler) handleStartCommand(message *tgbotapi.Message) error {
	greeting := "Welcome to using this bot. You can control the bot using these commands:\n\n"
	for _, c := range h.commands {
		greeting += fmt.Sprintf("/%s - %s\n", c.command, c.description)
	}

	greeting = strings.TrimRight(greeting, "\n")
	return h.sendText(h.chatId, greeting)
}

func (h *handler) handleChatIdCommand(message *tgbotapi.Message) error {
	return h.sendText(message.Chat.ID, fmt.Sprintf("%d", message.Chat.ID))
}

func (h *handler) handleSwitchSlotCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	simSlots, err := h.modem.GetSimSlots()
	if err != nil {
		return err
	}

	currentActivatedSlot, err := h.modem.GetPrimarySimSlot()
	if err != nil {
		return err
	}

	buttons := []tgbotapi.InlineKeyboardButton{}
	for slotIndex := range simSlots {
		if uint32(slotIndex)+1 == currentActivatedSlot {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Slot %d (active)", slotIndex+1), fmt.Sprintf("switchsimslot:%d", slotIndex+1)))
		} else {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Slot %d", slotIndex+1), fmt.Sprintf("switchsimslot:%d", slotIndex+1)))
		}
	}

	msg := tgbotapi.NewMessage(h.chatId, "*Which SIM slot do you want to use?*\nThis action may take some time\\.")
	msg.ParseMode = "markdownV2"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg.ReplyMarkup = keyboard
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) handleSwitchSlotCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	simSlot, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return err
	}

	err = h.modem.SetPrimarySimSlot(uint32(simSlot))
	if err != nil {
		return err
	}

	if _, err = h.tgbot.Request(tgbotapi.NewCallback(callback.ID, "Success!")); err != nil {
		return err
	}
	return h.sendText(callback.Message.Chat.ID, "Success! SIM slot has been changed.")
}

func (h *handler) handleSimCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	slot, _ := h.modem.GetPrimarySimSlot()
	operator, _ := h.modem.GetOperatorName()
	iccid, _ := h.modem.GetIccid()
	imei, _ := h.modem.GetImei()
	signalQuality, _ := h.modem.GetSignalQuality()
	return h.sendText(message.Chat.ID, fmt.Sprintf("SIM Slot: %d\nOperator: %s\nICCID: %s\nIMEI: %s\nSignal Quality: %d%s", slot, operator, iccid, imei, signalQuality, "%"))
}

func (h *handler) handleUSSDCommand(message *tgbotapi.Message) error {
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

func (h *handler) handleUSSDRespondCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 1 {
		return errors.New("invalid arguments")
	}

	reply, err := h.modem.RespondUSSDCommand(arguments[0])
	if err != nil {
		return err
	}
	return h.sendText(message.Chat.ID, reply)
}

func (h *handler) handleSendSmsCommand(message *tgbotapi.Message) error {
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

func (h *handler) checkChatId(chatId int64) error {
	if h.chatId == chatId {
		return nil
	}
	return errors.New("chat id does not match")
}

func (h *handler) sendText(chatId int64, message string) error {
	if _, err := h.tgbot.Send(tgbotapi.NewMessage(chatId, message)); err != nil {
		return err
	}
	return nil
}
