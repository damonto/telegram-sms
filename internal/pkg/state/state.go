package state

import (
	"sync"

	"gopkg.in/telebot.v3"
)

type State interface {
	States(stages map[string]telebot.HandlerFunc)
	Next(next string)
}

type state struct {
	chatId int64
	states map[string]telebot.HandlerFunc
	next   string
}

type StateManager struct {
	bot    *telebot.Bot
	mutex  sync.Mutex
	states map[int64]*state
}

func NewState(bot *telebot.Bot) *StateManager {
	s := &StateManager{
		bot:    bot,
		states: make(map[int64]*state, 10),
	}
	s.handleText()
	return s
}

func (c *StateManager) handleText() {
	c.bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		if state, ok := c.states[ctx.Chat().ID]; ok {
			if step, ok := state.states[state.next]; ok {
				return step(ctx)
			}
		}
		return nil
	})
}

func (s *StateManager) New(ctx telebot.Context) State {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	state := &state{
		chatId: ctx.Chat().ID,
		states: make(map[string]telebot.HandlerFunc),
	}
	s.states[ctx.Chat().ID] = state
	return state
}

func (s *StateManager) Done(ctx telebot.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.states, ctx.Chat().ID)
}

func (c *state) States(states map[string]telebot.HandlerFunc) {
	c.states = states
}

func (c *state) Next(next string) {
	c.next = next
}
