package modem

import (
	"errors"
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
