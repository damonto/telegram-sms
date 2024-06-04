package config

import "errors"

type Config struct {
	BotToken     string
	AdminId      int64
	IseUICC      bool
	DataDir      string
	LpacVersion  string
	DontDownload bool
	Verbose      bool
}

var C = &Config{}

var (
	ErrBotTokenRequired = errors.New("bot token is required")
	ErrAdminIdRequired  = errors.New("admin id is required")
	ErrDataDirRequired  = errors.New("data dir is required")
)

func (c *Config) IsValid() error {
	if c.BotToken == "" {
		return ErrBotTokenRequired
	}
	if c.IseUICC && c.DataDir == "" {
		return ErrDataDirRequired
	}
	return nil
}
