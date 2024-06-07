package config

import "errors"

type Config struct {
	BotToken     string
	AdminId      int64
	IsEuicc      bool
	Dir          string
	Version      string
	DontDownload bool
	Verbose      bool
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
