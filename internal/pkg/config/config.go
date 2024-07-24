package config

import (
	"errors"
)

const (
	APDUDriverAT  string = "at"
	APDUDriverQMI string = "qmi"
)

type Config struct {
	BotToken     string
	AdminId      int64
	Dir          string
	APDUDriver   string
	Version      string
	DontDownload bool
	Verbose      bool
}

var C = &Config{}

var (
	ErrBotTokenRequired      = errors.New("bot token is required")
	ErrUnsupportedAPDUDriver = errors.New("unsupported apdu driver")
)

func (c *Config) IsValid() error {
	if c.BotToken == "" {
		return ErrBotTokenRequired
	}
	if c.APDUDriver != APDUDriverAT && c.APDUDriver != APDUDriverQMI {
		return ErrUnsupportedAPDUDriver
	}
	return nil
}
