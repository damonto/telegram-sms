package conversation

import (
	"log/slog"

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
type converstations struct {
	bot           *telebot.Bot
	conversations map[int64]*conversation
}

var converstationsInstance *converstations

func NewConversation(bot *telebot.Bot) *converstations {
	converstationsInstance = &converstations{
		bot:           bot,
		conversations: make(map[int64]*conversation),
	}
	converstationsInstance.handleText()
	return converstationsInstance
}

func (c *converstations) handleText() {
	c.bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		if conv, ok := c.conversations[ctx.Sender().ID]; ok {
			if step, ok := conv.steps[conv.next]; ok {
				return step(ctx)
			}
			slog.Error("step not found", "step", conv.next)
		}
		return nil
	})
}

func New(ctx telebot.Context) Conversation {
	conversation := &conversation{
		chatId: ctx.Sender().ID,
		steps:  make(map[string]telebot.HandlerFunc),
	}
	converstationsInstance.conversations[ctx.Sender().ID] = conversation
	return conversation
}

func (c *conversation) Steps(steps map[string]telebot.HandlerFunc) {
	c.steps = steps
}

func (c *conversation) Next(next string) {
	c.next = next
}

func (c *conversation) Done() {
	delete(converstationsInstance.conversations, c.chatId)
}
