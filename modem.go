package main

import (
	"errors"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/exp/slog"
)

type SMSSubscriber = func(sms modemmanager.Sms)

type Modem interface {
	GetIccid() (string, error)
	GetOperatorName() (string, error)
	GetImei() (string, error)
	GetSignalQuality() (uint32, error)
	GetPrimarySimSlot() (uint32, error)
	SetPrimarySimSlot(simSlot uint32) error
	GetSimSlots() ([]dbus.ObjectPath, error)
	RunUSSDCommand(command string) (string, error)
	RespondUSSDCommand(response string) (string, error)
	SubscribeSMS(callback SMSSubscriber) error
	SendSMS(number, message string) error
}

type mm struct {
	mmgr               modemmanager.ModemManager
	modem              modemmanager.Modem
	smsResubscribeChan chan struct{}
}

var (
	modemPropertySimSlots       = "org.freedesktop.ModemManager1.Modem.SimSlots"
	modemPropertyPrimarySimSlot = "org.freedesktop.ModemManager1.Modem.PrimarySimSlot"
	modemSetPrimarySimSlot      = "org.freedesktop.ModemManager1.Modem.SetPrimarySimSlot"
)

func NewModem() (Modem, error) {
	var err error
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}
	mmgr.SetLogging(modemmanager.MMLoggingLevelWarning)
	m := &mm{
		mmgr:               mmgr,
		smsResubscribeChan: make(chan struct{}, 1),
	}
	m.modem, err = m.initModemManager()
	return m, err
}

func (m *mm) initModemManager() (modemmanager.Modem, error) {
	modems, err := m.mmgr.GetModems()
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
	return modem, err
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

func (m *mm) GetOperatorName() (string, error) {
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

func (m *mm) GetSimSlots() ([]dbus.ObjectPath, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), modemPropertySimSlots)
	if err != nil {
		return nil, err
	}

	simSlots := []dbus.ObjectPath{}
	for _, slot := range prop.Value().([]dbus.ObjectPath) {
		simSlots = append(simSlots, slot)
	}
	return simSlots, err
}

func (m *mm) GetPrimarySimSlot() (uint32, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), modemPropertyPrimarySimSlot)
	if err != nil {
		return 0, err
	}
	return prop.Value().(uint32), err
}

func (m *mm) SetPrimarySimSlot(simSlot uint32) error {
	primarySlot, err := m.GetPrimarySimSlot()
	if err != nil {
		return err
	}

	if primarySlot == simSlot {
		slog.Info("the given sim slot has already activated", "slot", simSlot)
		return nil
	}

	if err := m.callMethod(m.modem.GetObjectPath(), modemSetPrimarySimSlot, uint32(simSlot)); err != nil {
		return err
	}

	for {
		time.Sleep(5 * time.Second)
		modems, err := m.mmgr.GetModems()
		if err != nil {
			return err
		}

		if len(modems) > 0 {
			modem := modems[0]
			if modem.GetObjectPath() != m.modem.GetObjectPath() {
				slog.Info("new modem detected", "path", modem.GetObjectPath())

				state, err := modem.GetState()
				if err != nil {
					return err
				}
				if state == modemmanager.MmModemStateDisabled {
					if err := modem.Enable(); err != nil {
						slog.Error("failed to enable modem", "error", err)
						return err
					}
				}
				m.modem = modem
				m.smsResubscribeChan <- struct{}{}
				break
			}
		}
	}

	return nil
}

func (m *mm) callMethod(objectPath dbus.ObjectPath, method string, args ...interface{}) error {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return err
	}

	obj := dbusConn.Object(modemmanager.ModemManagerInterface, objectPath)
	return obj.Call(method, 0, args...).Err
}

func (m *mm) getProperty(objectPath dbus.ObjectPath, property string) (dbus.Variant, error) {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return dbus.Variant{}, err
	}

	obj := dbusConn.Object(modemmanager.ModemManagerInterface, objectPath)
	return obj.GetProperty(property)
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

func (m *mm) RespondUSSDCommand(response string) (string, error) {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	return ussd.Respond(response)
}

func (m *mm) SendSMS(number, message string) error {
	messaging, err := m.modem.GetMessaging()
	if err != nil {
		return err
	}

	sms, err := messaging.CreateSms(number, message)

	return sms.Send()
}

func (m *mm) SubscribeSMS(subscriber SMSSubscriber) error {
Subscribe:
	messaging, err := m.modem.GetMessaging()
	if err != nil {
		slog.Error("failed to get messaging", "error", err)
	}

	for {
		select {
		case smsSignal := <-messaging.SubscribeAdded():
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
						subscriber(sms)
						break
					}
				}
			}

			if state == modemmanager.MmSmsStateReceived {
				subscriber(sms)
			} else {
				continue
			}
		case <-m.smsResubscribeChan:
			slog.Info("SIM slot has been changed, unsubscribe and subscribe again")
			messaging.Unsubscribe()
			goto Subscribe
		}
	}
}

func (m *mm) parseSMSDatetime(path dbus.ObjectPath) (string, error) {
	prop, err := m.getProperty(path, modemmanager.SmsPropertyTimestamp)
	if err != nil {
		return "", err
	}

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
