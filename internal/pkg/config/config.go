package config

import (
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/damonto/euicc-go/driver"
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

type AID string

var supportedAIDs = []string{"sgp22", "5ber", "esimme"}

var AIDs = map[string][]byte{
	"sgp22":  driver.SGP22AID,
	"5ber":   []byte{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x05, 0x05, 0x00},
	"esimme": []byte{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0x00, 0x00, 0x00, 0x89, 0x00, 0x00, 0x00, 0x03, 0x00},
}

func (a *AID) Set(value string) error {
	if !slices.Contains(supportedAIDs, value) {
		return errors.New("unknown eUICC")
	}
	*a = AID(value)
	return nil
}

func (a *AID) String() string { return string(*a) }

func (a *AID) UnmarshalBinary() ([]byte, error) {
	if aid, ok := AIDs[string(*a)]; ok {
		return aid, nil
	}
	return nil, errors.New("unknown AID")
}

type Config struct {
	BotToken string
	AdminId  AdminId
	AID      AID
	Endpoint string
	Slowdown bool
	Verbose  bool
}

var C = &Config{AID: "sgp22"}

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
