package modem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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
		return NewCRSMUpdater(at).Update(hasPrefix, name, number)
	}
	return NewCSIMUpdater(at).Update(hasPrefix, name, number)
}

type Updater interface {
	Update(hasPrefix bool, name []byte, number []byte) error
}

type CRSMUpdater struct{ at *AT }

func NewCRSMUpdater(at *AT) Updater {
	return &CRSMUpdater{at: at}
}

func (u *CRSMUpdater) Update(hasPrefix bool, name []byte, number []byte) error {
	n, err := u.readLength()
	if err != nil {
		return err
	}
	if len(name) > n-14 {
		return errors.New("name is too long")
	}
	var data []byte
	data = append(data, paddingRight(name, n-14)...)
	data = append(data, append(
		[]byte{byte(len(number)), util.If(hasPrefix, byte(0x91), byte(0x81))},
		paddingRight(number, 12)...,
	)...)
	_, err = u.run(fmt.Sprintf("220,28480,1,4,%d,\"%X\"", n, data))
	return err
}

func (u *CRSMUpdater) readLength() (int, error) {
	r, err := u.run("192,28480")
	if err != nil {
		return 0, err
	}
	data := searchFcpContent(r, 0x82)
	len := int(data[4])*256 + int(data[5])
	return len, nil
}

func (u *CRSMUpdater) run(command string) ([]byte, error) {
	command = fmt.Sprintf("AT+CRSM=%s", command)
	slog.Debug("[CRSM] MSISDN Sending", "command", command)
	response, err := u.at.Run(command)
	slog.Debug("[CRSM] MSISDN Received", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	return u.sw(response)
}

func (u *CRSMUpdater) sw(sw string) ([]byte, error) {
	if !strings.Contains(sw, "+CRSM: 144") {
		return nil, fmt.Errorf("unexpected response: %s", sw)
	}
	data := strings.Replace(sw, "+CRSM: 144,0,", "", 1)
	return hex.DecodeString(data[1 : len(data)-1])
}

type CSIMUpdater struct{ at *AT }

func NewCSIMUpdater(at *AT) Updater {
	return &CSIMUpdater{at: at}
}

func (u *CSIMUpdater) Update(hasPrefix bool, name []byte, number []byte) error {
	n, err := u.selectFile()
	if err != nil {
		return err
	}
	if len(name) > n-14 {
		return errors.New("name is too long")
	}
	var data []byte
	data = append(data, paddingRight(name, n-14)...)
	data = append(data, append(
		[]byte{byte(len(number)), util.If(hasPrefix, byte(0x91), byte(0x81))},
		paddingRight(number, 12)...,
	)...)
	command := append([]byte{0x00, 0xDC, 0x01, 0x04}, byte(len(data)))
	command = append(command, data...)
	_, err = u.run(command)
	return err
}

func (u *CSIMUpdater) selectFile() (int, error) {
	r, err := u.run([]byte{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40})
	if err != nil {
		return 0, err
	}
	data := searchFcpContent(r, 0x82)
	len := int(data[4])*256 + int(data[5])
	// return u.run([]byte{0x00, 0xB2, 0x01, 0x04, byte(len)})
	return len, nil
}

func (u *CSIMUpdater) command(command []byte) string {
	cmd := fmt.Sprintf("%X", command)
	return fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)
}

func (u *CSIMUpdater) run(command []byte) ([]byte, error) {
	slog.Debug("[CSIM] MSISDN Sending command", "command", u.command(command))
	response, err := u.at.Run(u.command(command))
	slog.Debug("[CSIM] MSISDN Received response", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	sw, err := u.sw(response)
	if err != nil {
		return nil, err
	}
	if sw[0] != 0x61 && sw[len(sw)-2] != 0x90 {
		return sw, fmt.Errorf("unexpected response: %s", sw)
	}
	if sw[0] == 0x61 {
		return u.read(sw[1:])
	}
	return sw, nil
}

func (u *CSIMUpdater) read(length []byte) ([]byte, error) {
	command := append([]byte{0x00, 0xC0, 0x00, 0x00}, length...)
	return u.run(command)
}

func (u *CSIMUpdater) sw(sw string) ([]byte, error) {
	lastIdx := strings.LastIndex(sw, ",")
	if lastIdx == -1 {
		return nil, errors.New("invalid response")
	}
	return hex.DecodeString(sw[lastIdx+2 : len(sw)-1])
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

func searchFcpContent(bs []byte, tag byte) []byte {
	bs = bs[2:]
	for len(bs) > 0 {
		n := int(bs[1])
		if bs[0] == tag {
			return bs[:2+n]
		}
		bs = bs[2+n:]
	}
	return nil
}
