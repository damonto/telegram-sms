package conversation

import (
	"log/slog"
	"sync"

	"gopkg.in/telebot.v3"
)

type Conversation interface {
	Steps(steps map[string]telebot.HandlerFunc)
	Next(next string)
	Done()
}

type conversation struct {
	chatId int64
	steps  map[string]telebot.HandlerFunc
	next   string
}
type conversations struct {
	bot           *telebot.Bot
	mutex         sync.Mutex
	conversations map[int64]*conversation
}

var conversationsInstance *conversations

func NewConversation(bot *telebot.Bot) *conversations {
	conversationsInstance = &conversations{
		bot:           bot,
		conversations: make(map[int64]*conversation),
	}
	conversationsInstance.handleText()
	return conversationsInstance
}

func (c *conversations) handleText() {
	c.bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		if conv, ok := c.conversations[ctx.Chat().ID]; ok {
			if step, ok := conv.steps[conv.next]; ok {
				return step(ctx)
			}
			slog.Error("step not found", "step", conv.next)
		}
		return nil
	})
}

func New(ctx telebot.Context) Conversation {
	conversationsInstance.mutex.Lock()
	defer conversationsInstance.mutex.Unlock()
	conversation := &conversation{
		chatId: ctx.Chat().ID,
		steps:  make(map[string]telebot.HandlerFunc),
	}
	conversationsInstance.conversations[ctx.Chat().ID] = conversation
	return conversation
}

func (c *conversation) Steps(steps map[string]telebot.HandlerFunc) {
	c.steps = steps
}

func (c *conversation) Next(next string) {
	c.next = next
}

func (c *conversation) Done() {
	conversationsInstance.mutex.Lock()
	defer conversationsInstance.mutex.Unlock()
	delete(conversationsInstance.conversations, c.chatId)
}
