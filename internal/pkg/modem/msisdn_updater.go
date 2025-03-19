package modem

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/damonto/telegram-sms/internal/pkg/util"
)

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

type CRSMRunner struct{ at *AT }

func NewCRSMRunner(at *AT) MSISDNCommandRunner { return &CRSMRunner{at: at} }

func (r *CRSMRunner) Select() ([]byte, error) {
	sw, err := r.run("192,28480")
	if err != nil {
		return nil, err
	}
	return sw, nil
}

func (r *CRSMRunner) Run(data []byte) error {
	_, err := r.run(fmt.Sprintf("220,28480,1,4,%d,\"%X\"", len(data), data))
	return err
}

func (r *CRSMRunner) run(command string) ([]byte, error) {
	command = fmt.Sprintf("AT+CRSM=%s", command)
	slog.Debug("[CRSM] MSISDN Sending", "command", command)
	response, err := r.at.Run(command)
	slog.Debug("[CRSM] MSISDN Received", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	return r.sw(response)
}

func (r *CRSMRunner) sw(sw string) ([]byte, error) {
	if !strings.Contains(sw, "+CRSM: 144") {
		return nil, fmt.Errorf("unexpected response: %s", sw)
	}
	data := strings.Replace(sw, "+CRSM: 144,0,", "", 1)
	return hex.DecodeString(data[1 : len(data)-1])
}

// endregion

// region CSIMRunner

type CSIMRunner struct{ at *AT }

func NewCSIMRunner(at *AT) MSISDNCommandRunner { return &CSIMRunner{at: at} }

func (r *CSIMRunner) Run(data []byte) error {
	_, err := r.run(append([]byte{0x00, 0xDC, 0x01, 0x04, byte(len(data))}, data...))
	return err
}

func (r *CSIMRunner) Select() ([]byte, error) {
	return r.run([]byte{0x00, 0xA4, 0x08, 0x04, 0x04, 0x7F, 0xFF, 0x6F, 0x40})
}

func (r *CSIMRunner) command(command []byte) string {
	cmd := fmt.Sprintf("%X", command)
	return fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)
}

func (r *CSIMRunner) run(command []byte) ([]byte, error) {
	slog.Debug("[CSIM] MSISDN Sending", "command", r.command(command))
	response, err := r.at.Run(r.command(command))
	slog.Debug("[CSIM] MSISDN Received", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	sw, err := r.sw(response)
	if err != nil {
		return nil, err
	}
	if sw[0] != 0x61 && sw[len(sw)-2] != 0x90 {
		return sw, fmt.Errorf("unexpected response: %s", sw)
	}
	if sw[0] == 0x61 {
		return r.read(sw[1:])
	}
	return sw, nil
}

func (r *CSIMRunner) read(length []byte) ([]byte, error) {
	return r.run(append([]byte{0x00, 0xC0, 0x00, 0x00}, length...))
}

func (r *CSIMRunner) sw(sw string) ([]byte, error) {
	lastIdx := strings.LastIndex(sw, ",")
	if lastIdx == -1 {
		return nil, errors.New("invalid response")
	}
	return hex.DecodeString(sw[lastIdx+2 : len(sw)-1])
}

// endregion
