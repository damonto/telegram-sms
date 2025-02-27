package state

import (
	"log/slog"
	"sync"

	"github.com/damonto/telegram-sms/internal/app/middleware"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type Handler interface {
	Handle() th.Handler
	HandleMessage(ctx *th.Context, message telego.Message, state *ChatState) error
	HandleCallbackQuery(ctx *th.Context, query telego.CallbackQuery, state *ChatState) error
}

type State string

type ChatState struct {
	Handler Handler
	State   State
	Value   any
}

type StateManager struct {
	mutex  sync.Mutex
	states map[int64]*ChatState
}

var M *StateManager

func NewStateManager(handler *th.BotHandler) *StateManager {
	M = &StateManager{
		states: make(map[int64]*ChatState, 16),
	}
	return M
}

func (m *StateManager) RegisterCallback(handler *th.BotHandler) {
	handler.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {
		slog.Debug("Got callback query", "query", query.Data)
		state, ok := m.Get(query.Message.GetChat().ID)
		if !ok {
			slog.Debug("No state found", "chatID", query.From.ID, "query", query.Data)
			return nil
		}
		ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
		return state.Handler.HandleCallbackQuery(ctx, query, state)
	}, th.Not(th.CallbackDataPrefix(middleware.CallbackQueryAskModemPrefix)))
}

func (m *StateManager) RegisterMessage(handler *th.BotHandler) {
	handler.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		state, ok := m.Get(message.Chat.ID)
		if !ok {
			slog.Debug("No state found", "chatID", message.Chat.ID, "message", message.Text)
			return nil
		}
		return state.Handler.HandleMessage(ctx, message, state)
	}, th.Any())
}

func (m *StateManager) Enter(chatID int64, state *ChatState) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.states[chatID] = state
}

func (m *StateManager) Current(chatId int64, current State) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	state, ok := m.states[chatId]
	if !ok {
		return
	}
	state.State = current
}

func (m *StateManager) Get(chatID int64) (*ChatState, bool) {
	state, ok := m.states[chatID]
	return state, ok
}

func (m *StateManager) Exit(chatID int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.states, chatID)
}
