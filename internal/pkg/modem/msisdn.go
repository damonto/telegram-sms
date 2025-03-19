package modem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"regexp"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/util"
)

func (m *Modem) SetMSISDN(name string, number string) error {
	port, err := m.Port(ModemPortTypeAt)
	if err != nil {
		return err
	}
	at, err := NewAT(port.Device)
	if err != nil {
		return err
	}
	defer at.Close()
	if len(name) > 30 {
		return errors.New("the name can be at most 30 characters")
	}
	regexp, err := regexp.Compile(`^\+?[0-9]{1,15}$`)
	if err != nil {
		return err
	}
	if !regexp.MatchString(number) {
		return errors.New("invalid phone number")
	}
	nb, err := binaryCodedDecimalEncode(strings.TrimPrefix(number, "+"))
	if err != nil {
		return err
	}
	return m.updateMSISDN(at, strings.HasPrefix(number, "+"), []byte(name), nb)
}

func (m *Modem) updateMSISDN(at *AT, hasPrefix bool, name []byte, number []byte) error {
	if !at.Support("AT+CRSM=?") && !at.Support("AT+CSIM=?") {
		return errors.New("modem does not support updating MSISDN")
	}
	if at.Support("AT+CRSM=?") {
		return m.updateViaCRSM(at, hasPrefix, name, number)
	}
	return m.updateViaCSIM(at, hasPrefix, name, number)
}

func (m *Modem) updateViaCRSM(at *AT, hasPrefix bool, name []byte, number []byte) error {
	command, err := buildCRSMCommand(at, hasPrefix, name, number)
	if err != nil {
		return err
	}
	slog.Debug("[CRSIM] MSISDN Sending command", "command", command)
	sw, err := at.Run(command)
	slog.Debug("[CRSIM] MSISDN Received response", "response", sw, "error", err)
	if err != nil {
		return err
	}
	if !strings.Contains(sw, "+CRSM: 144,0") {
		return fmt.Errorf("failed to update MSISDN. SW: %s", sw)
	}
	return nil
}

func buildCRSMCommand(at *AT, hasPrefix bool, name []byte, number []byte) (string, error) {
	output, err := at.Run("AT+CRSM=178,28480,1,4,0")
	if err != nil {
		return "", err
	}
	valueLen := len(strings.Replace(strings.Replace(output, "+CRSM: 144,0,", "", 1), "\"", "", -1)) / 2
	number = append([]byte{byte(len(number)), util.If(hasPrefix, byte(0x91), byte(0x81))}, paddingRight(number, 13)...)
	cmd := fmt.Sprintf("%X", append(
		paddingRight(name, valueLen-len(number)),
		number...,
	))
	return fmt.Sprintf("AT+CRSM=220,28480,1,4,%d,\"%s\"", valueLen, cmd), nil
}

func (m *Modem) updateViaCSIM(at *AT, hasPrefix bool, name []byte, number []byte) error {
	for command := range buildCSIMCommands(hasPrefix, name, number) {
		slog.Debug("[CSIM] MSISDN Sending command", "command", command)
		sw, err := at.Run(command)
		slog.Debug("[CSIM] MSISDN Received response", "response", sw, "error", err)
		if err != nil {
			return err
		}
		sw = sw[len(sw)-5:]
		if sw[0:2] != "90" && sw[0:2] != "61" {
			return fmt.Errorf("failed to update MSISDN. SW: %s", sw)
		}
	}
	return nil
}

func buildCSIMCommands(hasPrefix bool, name []byte, number []byte) iter.Seq[string] {
	commands := [][]byte{
		{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40},
		{0x00, 0xDC, 0x01, 0x04, 0x1E}, // 0x01: We only need to update the first record.
	}
	commands[1] = append(commands[1], paddingRight(name, 15)...)
	commands[1] = append(commands[1], append(
		[]byte{byte(len(number)), util.If(hasPrefix, byte(0x91), byte(0x81))},
		paddingRight(number, 13)...,
	)...)
	return func(yield func(string) bool) {
		for _, command := range commands {
			cmd := fmt.Sprintf("%X", command)
			if !yield(fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)) {
				return
			}
		}
	}
}

func binaryCodedDecimalEncode(value string) ([]byte, error) {
	for _, r := range value {
		if (r < '0' || r > '9') && !(r == 'f' || r == 'F') {
			return nil, errors.New("invalid value")
		}
	}
	if len(value)%2 != 0 {
		value += "F"
	}
	id, _ := hex.DecodeString(value)
	for index := range id {
		id[index] = id[index]>>4 | id[index]<<4
	}
	return id, nil
}

func paddingRight(value []byte, length int) []byte {
	if len(value) >= length {
		return value
	}
	return append(value, bytes.Repeat([]byte{0xFF}, length-len(value))...)
}
