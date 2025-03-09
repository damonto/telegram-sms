package config

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
)

type AdminId []string

func (a *AdminId) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func (a *AdminId) String() string {
	return strings.Join(*a, ",")
}

func (a *AdminId) UnmarshalInt64() []int64 {
	var ids []int64
	for _, id := range *a {
		id, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Failed to convert admin id to int64", "id", id, "error", err)
			continue
		}
		ids = append(ids, int64(id))
	}
	return ids
}

type Config struct {
	BotToken string
	AdminId  AdminId
	Endpoint string
	Slowdown bool
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
