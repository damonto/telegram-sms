package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/damonto/telegram-sms/esim"
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
	chatId      int64
	isEuicc     bool
	tgbot       *tgbotapi.BotAPI
	modem       Modem
	botHandlers map[string]botHandler
	messages    map[int]*tgbotapi.Message
}

type botHandler struct {
	command     string
	description string
	handler     HandlerFunc
	callback    CallbackFunc
}

func NewHandler(chatId int64, isEuicc bool, tgbot *tgbotapi.BotAPI, modem Modem) Handler {
	return &handler{
		chatId:   chatId,
		isEuicc:  isEuicc,
		tgbot:    tgbot,
		modem:    modem,
		messages: make(map[int]*tgbotapi.Message, 1),
	}
}

func (h *handler) RegisterCommands() error {
	h.botHandlers = map[string]botHandler{
		"switchmodemcallback":   {callback: h.handleSwitchModemCallback},
		"chatid":                {command: "chatid", description: "Obtain your chat id", handler: h.handleChatIdCommand},
		"sim":                   {command: "sim", description: "Obtain SIM card properties", handler: h.handleSimCommand},
		"switchsimslot":         {command: "switchsimslot", description: "Switch to another SIM slot", handler: h.handleSwitchSlotCommand},
		"switchsimslotcallback": {callback: h.handleSwitchSlotCallback},
		"sms":                   {command: "sms", description: "Send an SMS to a phone number", handler: h.handleSendSmsCommand},
		"ussd":                  {command: "ussd", description: "Send a USSD command to your SIM card", handler: h.handleUSSDCommand},
		"ussdresponed":          {command: "ussdresponed", description: "Respond the last ussd command", handler: h.handleUSSDRespondCommand},
	}

	if h.isEuicc {
		h.botHandlers["esimprofiles"] = botHandler{command: "esimprofiles", description: "List installed eSIM profiles", handler: h.handleListEsimProfiles}
	}

	botCommands := []tgbotapi.BotCommand{}
	for _, c := range h.botHandlers {
		if c.command != "" {
			botCommands = append(botCommands, tgbotapi.BotCommand{
				Command:     c.command,
				Description: c.description,
			})
		}
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

	if command, ok := h.botHandlers[command]; ok {
		return command.handler(message)
	}
	return errors.New("command not found")
}

func (h *handler) HandleCallback(callback *tgbotapi.CallbackQuery) error {
	button := strings.Split(callback.Data, ":")

	if command, ok := h.botHandlers[button[0]]; ok {
		return command.callback(callback, strings.Join(button[1:], ":"))
	}
	return errors.New("command not found")
}

func (h *handler) handleStartCommand(message *tgbotapi.Message) error {
	welcomeMessage := "Welcome to using this bot. You can control the bot using these commands:\n\n"
	for _, c := range h.botHandlers {
		if c.command != "" {
			welcomeMessage += fmt.Sprintf("/%s - %s\n", c.command, c.description)
		}
	}

	welcomeMessage = strings.TrimRight(welcomeMessage, "\n")
	return h.sendText(h.chatId, welcomeMessage, message.MessageID)
}

func (h *handler) handleChatIdCommand(message *tgbotapi.Message) error {
	return h.sendText(message.Chat.ID, fmt.Sprintf("%d", message.Chat.ID), message.MessageID)
}

func (h *handler) handleSwitchSlotCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "switchsimslot"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

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
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Slot %d (active)", slotIndex+1), fmt.Sprintf("switchsimslotcallback:%d", slotIndex+1)))
		} else {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Slot %d", slotIndex+1), fmt.Sprintf("switchsimslotcallback:%d", slotIndex+1)))
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg := tgbotapi.NewMessage(h.chatId, "*Which SIM slot do you want to use?*\nThis action may take some time\\.")
	msg.ParseMode = "markdownV2"
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) chooseModem(message *tgbotapi.Message, command string) (bool, error) {
	if _, ok := h.messages[message.MessageID]; ok {
		slog.Info("modem already switched", "message-id", message.MessageID)
		return false, nil
	}

	h.messages[message.MessageID] = message
	modems, err := h.modem.ListModems()
	if err != nil {
		return false, err
	}

	if len(modems) > 1 {
		return true, h.sendSwitchModemReplyButtons(modems, message, command)
	}
	// If there is only one modem, use it
	for modemId, modem := range modems {
		slog.Info("using modem", "modem-id", modemId, "modem", modem)
		h.modem.Use(modemId)
	}
	return false, nil
}

func (h *handler) sendSwitchModemReplyButtons(modems map[string]string, message *tgbotapi.Message, command string) error {
	buttons := []tgbotapi.InlineKeyboardButton{}
	for modemId, modem := range modems {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(modem, fmt.Sprintf("switchmodemcallback:%s:%s:%d", modemId, command, message.MessageID)))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	msg := tgbotapi.NewMessage(h.chatId, "Which modem do you want to use?")
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) handleSwitchModemCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	arguments := strings.Split(value, ":")
	modemId, next := arguments[0], arguments[1]
	messageId, err := strconv.ParseInt(arguments[2], 10, 32)
	if err != nil {
		return err
	}

	h.modem.Use(modemId)
	slog.Info("switched modem", "modem-id", modemId)

	if _, err := h.tgbot.Request(tgbotapi.NewCallback(callback.ID, "Success!")); err != nil {
		return err
	}
	if message, ok := h.messages[int(messageId)]; ok {
		slog.Info("passing message to next handler", "message-id", messageId)
		return h.HandleCommand(next, message)
	}

	return h.HandleCommand(next, callback.Message)
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
	return h.sendText(callback.Message.Chat.ID, "Success! SIM slot has been changed.", callback.Message.MessageID)
}

func (h *handler) handleSimCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "sim"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	slot, _ := h.modem.GetPrimarySimSlot()
	operator, _ := h.modem.GetOperatorName()
	iccid, _ := h.modem.GetIccid()
	imei, _ := h.modem.GetImei()
	signalQuality, _ := h.modem.GetSignalQuality()

	if !h.isEuicc {
		return h.sendText(
			message.Chat.ID,
			fmt.Sprintf("SIM Slot: %d\nOperator Name: %s\nICCID: %s\nIMEI: %s\nSignal Quality: %d%s", slot, operator, iccid, imei, signalQuality, "%"),
			message.MessageID,
		)
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}

	esim := esim.New(device)
	profiles, err := esim.ListProfiles()
	if err != nil {
		return err
	}

	var profileName string
	for _, profile := range profiles {
		if profile.Iccid == iccid {
			if profile.ProfileNickname != "" {
				profileName = profile.ProfileNickname
			} else {
				profileName = profile.ProviderName
			}
		}
	}
	eid, err := esim.Eid()
	if err != nil {
		return err
	}

	return h.sendText(
		message.Chat.ID,
		fmt.Sprintf("SIM Slot: %d\nEID: %s\nOperator Name: %s\nProvider Name: %s\nICCID: %s\nIMEI: %s\nSignal Quality: %d%s", slot, eid, operator, profileName, iccid, imei, signalQuality, "%"),
		message.MessageID,
	)
}

func (h *handler) handleUSSDCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "ussd"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 1 {
		return errors.New("invalid arguments")
	}

	result, err := h.modem.RunUSSDCommand(arguments[0])
	if err != nil {
		return err
	}
	return h.sendText(message.Chat.ID, result, message.MessageID)
}

func (h *handler) handleUSSDRespondCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "ussdresponed"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 1 {
		return errors.New("invalid arguments")
	}

	reply, err := h.modem.RespondUSSDCommand(arguments[0])
	if err != nil {
		return err
	}
	return h.sendText(message.Chat.ID, reply, message.MessageID)
}

func (h *handler) handleSendSmsCommand(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "sms"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 2 {
		return errors.New("invalid arguments")
	}

	return h.modem.SendSMS(arguments[0], strings.Join(arguments[1:], " "))
}

func (h *handler) checkChatId(chatId int64) error {
	if h.chatId == chatId {
		return nil
	}
	return errors.New("chat id does not match")
}

func (h *handler) handleListEsimProfiles(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if yes, err := h.chooseModem(message, "esimprofiles"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	profiles, err := esim.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return h.sendText(message.Chat.ID, "No eSIM profiles found.", message.MessageID)
	}

	content := ""
	for _, profile := range profiles {
		state := "Disabled"
		if profile.State == 1 {
			state = "Enabled"
		}
		content += fmt.Sprintf("Profile Name: %s\nNickname: %s\nProvider Name: %s\nICCID: %s\nState: %s\n\n", profile.ProfileName, profile.ProfileNickname, profile.ProviderName, profile.Iccid, state)
	}

	return h.sendText(message.Chat.ID, content, message.MessageID)
}

func (h *handler) sendText(chatId int64, message string, messageId int) error {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ReplyToMessageID = messageId
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}
