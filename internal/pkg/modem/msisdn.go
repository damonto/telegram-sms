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
	regexp, err := regexp.Compile(`^\+?[0-9]{1,15}$`)
	if err != nil {
		return err
	}
	if !regexp.MatchString(number) {
		return errors.New("invalid phone number")
	}
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
	return m.updateMSISDN(at, strings.HasPrefix(number, "+"), nameBytes, numberBytes)
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
	command, err := m.CRSMCommand(at, hasPrefix, name, number)
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

func (m *Modem) CRSMCommand(at *AT, hasPrefix bool, name []byte, number []byte) (string, error) {
	output, err := at.Run("AT+CRSM=178,28480,1,4,0")
	if err != nil {
		return "", err
	}
	valueLen := len(strings.Replace(strings.Replace(output, "+CRSM: 144,0,", "", 1), "\"", "", -1)) / 2
	numberType := []byte{0x05, 0x81}
	if hasPrefix {
		numberType = []byte{0x07, 0x91}
	}
	number = append(numberType, paddingRight(number, 12)...)
	cmd := fmt.Sprintf("%X", append(
		paddingRight(name, valueLen-len(number)), // Name
		number...,                                // Phone Number
	))
	return fmt.Sprintf("AT+CRSM=220,28480,1,4,%d,\"%s\"", valueLen, cmd), nil
}

func (m *Modem) updateViaCSIM(at *AT, hasPrefix bool, name []byte, number []byte) error {
	for command := range m.CSIMCommands(hasPrefix, name, number) {
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

func (m *Modem) CSIMCommands(hasPrefix bool, name []byte, number []byte) iter.Seq[string] {
	commands := [][]byte{
		{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40},
		{0x00, 0xDC, 0x01, 0x04, 0x1E, 0x4C}, // 0x01: We only need to update the first record.
	}
	numberType := []byte{0x05, 0x81}
	if hasPrefix {
		numberType = []byte{0x07, 0x91}
	}
	commands[1] = append(commands[1], paddingRight(name, 15)...)
	commands[1] = append(commands[1], append(append(commands[1], numberType...), paddingRight(number, 12)...)...)
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
