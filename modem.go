package main

import (
	"errors"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/exp/slog"
)

type SMSReceviedFunc = func(sms modemmanager.Sms)

type Modem interface {
	GetIccid() (string, error)
	GetCarrier() (string, error)
	GetImei() (string, error)
	GetSignalQuality() (uint32, error)
	RunUSSDCommand(command string) (string, error)
	SMSRecevied(callback SMSReceviedFunc) error
	SendSMS(number, message string) error
}

type mm struct {
	modem modemmanager.Modem
}

func NewModem() (Modem, error) {
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}

	modems, err := mmgr.GetModems()
	if err != nil {
		return nil, err
	}

	if len(modems) == 0 {
		return nil, errors.New("no modems found")
	}

	// Only support one modem for now
	modem := modems[0]
	state, err := modem.GetState()
	if err != nil {
		return nil, err
	}
	if state == modemmanager.MmModemStateDisabled {
		if err := modem.Enable(); err != nil {
			slog.Error("failed to enable modem", "error", err)
			return nil, err
		}
	}

	return &mm{
		modem,
	}, nil
}

func (m *mm) GetIccid() (string, error) {
	sim, err := m.modem.GetSim()
	if err != nil {
		return "", err
	}

	return sim.GetSimIdentifier()
}

func (m *mm) GetImei() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	return threeGpp.GetImei()
}

func (m *mm) GetCarrier() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	return threeGpp.GetOperatorName()
}

func (m *mm) GetSignalQuality() (uint32, error) {
	percent, _, err := m.modem.GetSignalQuality()
	if err != nil {
		return 0, err
	}

	return percent, err
}

func (m *mm) RunUSSDCommand(command string) (string, error) {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}

	return ussd.Initiate(command)
}

func (m *mm) SendSMS(number, message string) error {
	messaging, err := m.modem.GetMessaging()
	if err != nil {
		return err
	}

	sms, err := messaging.CreateSms(number, message)
	return sms.Send()
}

func (m *mm) SMSRecevied(callback SMSReceviedFunc) error {
	messaging, err := m.modem.GetMessaging()
	if err != nil {
		slog.Error("failed to get messaging", "error", err)
	}

	for {
		smsSignal := <-messaging.SubscribeAdded()
		sms, _, err := messaging.ParseAdded(smsSignal)
		if err != nil {
			slog.Error("failed to parse SMS", "error", err)
			continue
		}

		state, err := sms.GetState()
		if err != nil {
			slog.Error("failed to get SMS state", "error", err)
			continue
		}

		if state == modemmanager.MmSmsStateReceiving {
			for {
				time.Sleep(1 * time.Second)
				if state, err := sms.GetState(); state == modemmanager.MmSmsStateReceived && err == nil {
					callback(sms)
					break
				}
			}
		}

		if state == modemmanager.MmSmsStateReceived {
			callback(sms)
		} else {
			continue
		}
	}
}

func (m *mm) parseSMSDatetime(path dbus.ObjectPath) (string, error) {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return "", err
	}

	obj := dbusConn.Object(modemmanager.ModemManagerInterface, path)
	prop, err := obj.GetProperty(modemmanager.SmsPropertyTimestamp)

	for _, f := range []string{
		"2006-01-02T15:04:05-07",
		time.RFC3339,
		time.RFC3339Nano,
	} {
		t, err := time.Parse(f, prop.Value().(string))
		if err == nil {
			return t.Format(time.DateTime), nil
		}
	}

	return time.Now().Format(time.DateTime), nil
}
