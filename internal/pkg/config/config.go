package config

import (
	"errors"
)

const (
	APDUDriverAT  string = "at"
	APDUDriverQMI string = "qmi"
)

type Config struct {
	BotToken string
	AdminId  int64
	Endpoint string
	Verbose  bool
}

var C = &Config{}

var (
	ErrBotTokenRequired = errors.New("bot token is required")
)

func (c *Config) IsValid() error {
	if c.BotToken == "" {
		return ErrBotTokenRequired
	}
	return nil
}
