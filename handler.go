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
type NextFunc = func(message *tgbotapi.Message, callback *tgbotapi.CallbackQuery, value string) error

type Handler interface {
	RegisterCommands() error
	HandleCommand(command string, message *tgbotapi.Message) error
	HandleCallback(callback *tgbotapi.CallbackQuery) error
	HandleRawMessage(message *tgbotapi.Message) error
}

type handler struct {
	chatId           int64
	isEuicc          bool
	tgbot            *tgbotapi.BotAPI
	modem            Modem
	botHandlers      map[string]botHandler
	triggeredMessage triggeredMessage
	nextAction       NextFunc
	messages         map[int]*tgbotapi.Message
}

type triggeredMessage struct {
	callback *tgbotapi.CallbackQuery
	value    string
}

type botHandler struct {
	command     string
	description string
	handler     HandlerFunc
	callback    CallbackFunc
}

func NewHandler(chatId int64, isEuicc bool, tgbot *tgbotapi.BotAPI, modem Modem) Handler {
	return &handler{
		chatId:           chatId,
		isEuicc:          isEuicc,
		tgbot:            tgbot,
		modem:            modem,
		messages:         make(map[int]*tgbotapi.Message, 1),
		triggeredMessage: triggeredMessage{},
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
		h.botHandlers["esimdownload"] = botHandler{command: "esimdownload", description: "Download a new eSIM profile", handler: h.handleEsimDownload}
		h.botHandlers["clickprofilecallback"] = botHandler{callback: h.handleClickProfileCallback}
		h.botHandlers["enableprofilecallback"] = botHandler{callback: h.handleEnableProfileCallback}
		h.botHandlers["disableprofilecallback"] = botHandler{callback: h.handleDisableProfileCallback}
		h.botHandlers["renameprofilecallback"] = botHandler{callback: h.handleRenameProfileCallback}
		h.botHandlers["deleteprofilecallback"] = botHandler{callback: h.handleDeleteProfileCallback}
		h.botHandlers["confirmdeleteprofilecallback"] = botHandler{callback: h.handleConfirmDeleteProfileCallback}
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
	return errors.New("callback not found")
}

func (h *handler) HandleRawMessage(message *tgbotapi.Message) error {
	if h.nextAction == nil {
		return errors.New("undefined next action")
	}

	return h.nextAction(message, h.triggeredMessage.callback, h.triggeredMessage.value)
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

	activatedSlot, err := h.modem.GetPrimarySimSlot()
	if err != nil {
		return err
	}

	buttons := []tgbotapi.InlineKeyboardButton{}
	for slotIndex := range simSlots {
		if uint32(slotIndex)+1 == activatedSlot {
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
	buttons := [][]tgbotapi.InlineKeyboardButton{}
	for modemId, modem := range modems {
		button := tgbotapi.NewInlineKeyboardButtonData(modem, fmt.Sprintf("switchmodemcallback:%s:%s:%d", modemId, command, message.MessageID))
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
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

	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Disable", "disableprofilecallback:"+iccid),
		tgbotapi.NewInlineKeyboardButtonData("Rename", "renameprofilecallback:"+iccid),
		tgbotapi.NewInlineKeyboardButtonData("Delete", "deleteprofilecallback:"+iccid),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	content := fmt.Sprintf("SIM Slot: %d\nEID: %s\nOperator Name: %s\nProvider Name: %s\nICCID: %s\nIMEI: %s\nSignal Quality: %d%s", slot, eid, operator, profileName, iccid, imei, signalQuality, "%")
	msg := tgbotapi.NewMessage(h.chatId, EscapeText(content))
	msg.ParseMode = "markdownV2"
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
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

func (h *handler) handleEsimDownload(message *tgbotapi.Message) error {
	if err := h.checkChatId(message.Chat.ID); err != nil {
		return err
	}

	if message.CommandArguments() == "" {
		return errors.New("please send me the SM-DP+ address, activation code and confirmation code (optional)")
	}

	if yes, err := h.chooseModem(message, "esimdownload"); err != nil || yes {
		return err
	}
	delete(h.messages, message.MessageID)

	arguments := strings.Split(message.CommandArguments(), " ")
	if len(arguments) < 2 {
		return errors.New("invalid arguments")
	}

	confirmationCode := ""
	if len(arguments) == 3 {
		confirmationCode = arguments[2]
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	imei, _ := h.modem.GetImei()
	if err := esim.Download(arguments[0], arguments[1], confirmationCode, imei); err != nil {
		return err
	}

	return h.sendText(message.Chat.ID, "Congratulations! your new eSIM is ready!", message.MessageID)
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

	buttons := []tgbotapi.InlineKeyboardButton{}
	buttonRows := [][]tgbotapi.InlineKeyboardButton{}
	content := ""
	for _, profile := range profiles {
		state := "Disabled"
		if profile.State == 1 {
			state = "Enabled"
		}
		content += fmt.Sprintf("Profile Name: %s\nNickname: %s\nProvider Name: %s\nICCID: %s\nState: %s\n\n", profile.ProfileName, profile.ProfileNickname, profile.ProviderName, profile.Iccid, state)
		profileName := profile.ProfileName
		if profile.ProfileNickname != "" {
			profileName = profile.ProfileNickname
		}

		if profile.State == 1 {
			profileName = profileName + "(active)"
		}
		if len(buttons) == 2 {
			buttonRows = append(buttonRows, buttons)
			buttons = []tgbotapi.InlineKeyboardButton{}
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(profileName, fmt.Sprintf("clickprofilecallback:%s", profile.Iccid)))
	}

	if len(buttons) > 0 {
		buttonRows = append(buttonRows, buttons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
	msg := tgbotapi.NewMessage(h.chatId, EscapeText(content))
	msg.ParseMode = "markdownV2"
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) handleClickProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	e := esim.New(device)
	profiles, err := e.ListProfiles()
	if err != nil {
		return err
	}

	profile := esim.Profile{}
	for _, p := range profiles {
		if p.Iccid == value {
			profile = p
			break
		}
	}

	if profile.Iccid == "" {
		return errors.New("no profile founded")
	}

	buttons := []tgbotapi.InlineKeyboardButton{}
	if profile.State == 1 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Disable", "disableprofilecallback:"+profile.Iccid))
	} else {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Enable", "enableprofilecallback:"+profile.Iccid))
	}

	buttons = append(buttons,
		tgbotapi.NewInlineKeyboardButtonData("Rename", "renameprofilecallback:"+profile.Iccid),
		tgbotapi.NewInlineKeyboardButtonData("Delete", "deleteprofilecallback:"+profile.Iccid),
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
	state := "Enabled"
	if profile.State == 0 {
		state = "Disabled"
	}
	content := fmt.Sprintf("Profile Name: %s\nProvider Name: %s\nNickname: %s\nICCID: %s\nState: %s", profile.ProfileName, profile.ProviderName, profile.ProfileNickname, profile.Iccid, state)
	msg := tgbotapi.NewMessage(h.chatId, EscapeText(content))
	msg.ParseMode = "markdownV2"
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = callback.Message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) handleEnableProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	if err := esim.Enable(value); err != nil {
		return err
	}

	if err := h.modem.Reload(); err != nil {
		return err
	}

	_, err = h.tgbot.Request(tgbotapi.NewCallback(callback.ID, "Success!"))
	if err != nil {
		slog.Error("failed to send callback", "error", err)
	}
	return h.sendText(callback.Message.Chat.ID, "Success! The eSIM profile has been enabled.", callback.Message.MessageID)
}

func (h *handler) handleDisableProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	if err := esim.Disable(value); err != nil {
		return err
	}

	if err := h.modem.Reload(); err != nil {
		return err
	}

	_, err = h.tgbot.Request(tgbotapi.NewCallback(callback.ID, "Success!"))
	if err != nil {
		slog.Error("failed to send callback", "error", err)
	}
	return h.sendText(callback.Message.Chat.ID, "Success! The eSIM profile has been disabled.", callback.Message.MessageID)
}

func (h *handler) handleRenameProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	h.triggeredMessage = triggeredMessage{
		callback: callback,
		value:    value,
	}
	h.nextAction = h.renameProfile

	return h.sendText(callback.Message.Chat.ID, "Please enter a new name for the eSIM.", callback.Message.MessageID)
}

func (h *handler) renameProfile(message *tgbotapi.Message, callback *tgbotapi.CallbackQuery, value string) error {
	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	if err := esim.Rename(value, message.Text); err != nil {
		return err
	}

	// clean
	h.triggeredMessage = triggeredMessage{}
	h.nextAction = nil

	return h.sendText(message.Chat.ID, "The profile name has been changed.", callback.Message.MessageID)
}

func (h *handler) handleDeleteProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Yes", "confirmdeleteprofilecallback:"+value)))
	msg := tgbotapi.NewMessage(h.chatId, "Are you sure you want to delete this profile?")
	msg.ParseMode = "markdownV2"
	msg.ReplyMarkup = keyboard
	msg.ReplyToMessageID = callback.Message.MessageID
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (h *handler) handleConfirmDeleteProfileCallback(callback *tgbotapi.CallbackQuery, value string) error {
	if err := h.checkChatId(callback.Message.Chat.ID); err != nil {
		return err
	}

	device, err := h.modem.GetAtDevice()
	if err != nil {
		return err
	}
	esim := esim.New(device)
	if err := esim.Delete(value); err != nil {
		return err
	}

	return h.sendText(callback.Message.Chat.ID, "The profile "+value+" has been deleted.", callback.Message.MessageID)
}

func (h *handler) sendText(chatId int64, message string, messageId int) error {
	msg := tgbotapi.NewMessage(chatId, message)
	msg.ReplyToMessageID = messageId
	if _, err := h.tgbot.Send(msg); err != nil {
		return err
	}
	return nil
}
