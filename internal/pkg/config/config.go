package config

import (
	"errors"
	"strconv"
	"strings"
)

type ChatId []string

func (c *ChatId) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func (c *ChatId) String() string {
	return strings.Join(*c, ",")
}

func (c *ChatId) ToInt64() []int64 {
	var ids []int64
	for _, id := range *c {
		id, err := strconv.Atoi(id)
		if err != nil {
			continue
		}
		ids = append(ids, int64(id))
	}
	return ids
}

type Config struct {
	BotToken string
	AdminId  ChatId
	Endpoint string
	Verbose  bool
}

var C = new(Config)

var (
	ErrBotTokenRequired = errors.New("bot token is required")
	ErrAdminIdRequired  = errors.New("admin id is required")
)

func (c *Config) IsValid() error {
	if c.BotToken == "" {
		return ErrBotTokenRequired
	}
	if len(c.AdminId) == 0 {
		return ErrAdminIdRequired
	}
	return nil
}
