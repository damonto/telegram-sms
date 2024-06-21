package state

import (
	"log/slog"
	"sync"

	"gopkg.in/telebot.v3"
)

type State interface {
	Stages(stages map[string]telebot.HandlerFunc)
	Next(next string)
}

type state struct {
	chatId int64
	stages map[string]telebot.HandlerFunc
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
			if step, ok := state.stages[state.next]; ok {
				return step(ctx)
			}
			slog.Error("stage not found", "stage", state.next)
		}
		return nil
	})
}

func (s *StateManager) New(ctx telebot.Context) State {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	state := &state{
		chatId: ctx.Chat().ID,
		stages: make(map[string]telebot.HandlerFunc),
	}
	s.states[ctx.Chat().ID] = state
	return state
}

func (s *StateManager) Done(ctx telebot.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.states, ctx.Chat().ID)
}

func (c *state) Stages(stages map[string]telebot.HandlerFunc) {
	c.stages = stages
}

func (c *state) Next(next string) {
	c.next = next
}
