package modem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"strings"
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
	defer at.f.Close()
	numberBytes, err := binaryCodedDecimalEncode(strings.TrimPrefix(number, "+"))
	if err != nil {
		return err
	}
	if len(numberBytes) > 12 {
		return errors.New("the phone number can be at most 12 characters")
	}
	nameBytes := []byte(name)
	if len(nameBytes) > 15 {
		return errors.New("the name can be at most 15 characters")
	}
	for command := range m.SetMSISDNCommands(strings.HasPrefix(number, "+"), nameBytes, numberBytes) {
		slog.Debug("[MSISDN] Sending command", "command", command)
		sw, err := at.Run(command)
		slog.Debug("[MSISDN] Received response", "response", sw, "error", err)
		if err != nil {
			return err
		}
		sw = sw[len(sw)-5:]
		if sw[0:2] != "90" && sw[0:2] != "61" {
			return errors.New("failed to set MSISDN: " + sw)
		}
	}
	return nil
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

func (m *Modem) SetMSISDNCommands(hasPrefix bool, name []byte, number []byte) iter.Seq[string] {
	commands := [][]byte{
		{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40}, // Select DF Telecom -> MSISDN
		{0x00, 0xDC, 0x01, 0x04, 0x1E, 0x4C},                   // Update Binary
	}
	numberType := []byte{0x05, 0x81}
	if hasPrefix {
		numberType = []byte{0x07, 0x91}
	}
	commands[1] = append(commands[1], paddingRight(name, 15)...)
	commands[1] = append(append(commands[1], numberType...), paddingRight(number, 12)...)
	return func(yield func(string) bool) {
		for _, command := range commands {
			cmd := fmt.Sprintf("%X", command)
			if !yield(fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)) {
				return
			}
		}
	}
}
