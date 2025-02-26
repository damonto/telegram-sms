package state

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type Handler interface {
	th.CallbackQueryHandler | th.InlineQueryHandler | th.MessageHandler
}

type State string

type ChatState[T Handler] struct {
	Handler T
	State   State
	Value   any
}

type StateManager map[telego.ChatID]struct {
	State any
}

func NewStateManager() StateManager {
	return make(StateManager)
}
