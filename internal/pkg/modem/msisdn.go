package modem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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
	return m.updateMSISDN(at, strings.HasPrefix(number, "+"), name, number)
}

func (m *Modem) updateMSISDN(at *AT, hasPrefix bool, name string, number string) error {
	if !at.Support("AT+CRSM=?") && !at.Support("AT+CSIM=?") {
		return errors.New("modem does not support updating MSISDN")
	}
	var runner MSISDNCommandRunner
	if at.Support("AT+CRSM=?") {
		runner = NewCRSMRunner(at)
	} else {
		runner = NewCSIMRunner(at)
	}
	return NewMSISDNUpdater(runner).Update(hasPrefix, name, number)
}

type MSISDNUpdater interface {
	Update(hasPrefix bool, name string, number string) error
}

type MSISDNCommandRunner interface {
	Run(data []byte) error
	Select() ([]byte, error)
}

type updater struct {
	runner MSISDNCommandRunner
}

func NewMSISDNUpdater(runner MSISDNCommandRunner) MSISDNUpdater {
	return &updater{runner: runner}
}

func (u *updater) Update(hasPrefix bool, name string, number string) error {
	n, err := u.len()
	if err != nil {
		return err
	}
	if len(name) > n-14 {
		return errors.New("name is too long")
	}
	nb, err := u.binaryCodedDecimalEncode(strings.TrimPrefix(number, "+"))
	if err != nil {
		return err
	}
	var data []byte
	data = append(data, u.paddingRight([]byte(name), n-14)...)
	data = append(data, append(
		[]byte{byte(len(nb) + 1), util.If(hasPrefix, byte(0x91), byte(0x81))},
		u.paddingRight(nb, 12)...,
	)...)
	return u.runner.Run(data)
}

func (u *updater) len() (int, error) {
	b, err := u.runner.Select()
	if err != nil {
		return 0, err
	}
	data := u.search(b, 0x82)
	if data == nil {
		return 0, fmt.Errorf("unexpected response: %X", b)
	}
	return int(data[4])<<8 + int(data[5]), nil
}

func (u *updater) binaryCodedDecimalEncode(value string) ([]byte, error) {
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

func (u *updater) paddingRight(value []byte, length int) []byte {
	if len(value) >= length {
		return value
	}
	return append(value, bytes.Repeat([]byte{0xFF}, length-len(value))...)
}

func (u *updater) search(bs []byte, tag byte) []byte {
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

// region CRSMRunner

type CRSMRunner struct {
	commander ATCommand
}

func NewCRSMRunner(at *AT) MSISDNCommandRunner { return &CRSMRunner{commander: NewCRSM(at)} }

func (r *CRSMRunner) Select() ([]byte, error) {
	command := CRSMCommand{Instruction: CRSMGetResponse, FileID: 0x6F40}
	return r.commander.Run(command.Bytes())
}

func (r *CRSMRunner) Run(data []byte) error {
	command := CRSMCommand{
		Instruction: CRSMUpdateRecord,
		FileID:      0x6F40,
		P1:          1,
		P2:          4,
		Data:        data,
	}
	_, err := r.commander.Run(command.Bytes())
	return err
}

// endregion

// region CSIMRunner

type CSIMRunner struct{ commander ATCommand }

func NewCSIMRunner(at *AT) MSISDNCommandRunner { return &CSIMRunner{commander: NewCSIM(at)} }

func (r *CSIMRunner) Run(data []byte) error {
	_, err := r.commander.Run(append([]byte{0x00, 0xDC, 0x01, 0x04, byte(len(data))}, data...))
	return err
}

func (r *CSIMRunner) Select() ([]byte, error) {
	return r.commander.Run([]byte{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40})
}

// endregion
