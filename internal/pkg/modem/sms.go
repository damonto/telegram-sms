package modem

import (
	"time"

	"github.com/godbus/dbus/v5"
)

const ModemSMSInterface = ModemManagerInterface + ".Sms"

type SMS struct {
	objectPath dbus.ObjectPath
	State      SMSState
	Number     string
	Text       string
	Timestamp  time.Time
}

func (m *Modem) RetrieveSMS(objectPath dbus.ObjectPath) (*SMS, error) {
	dbusObject, err := m.privateDbusObject(objectPath)
	if err != nil {
		return nil, err
	}
	sms := &SMS{objectPath: objectPath}
	variant, err := dbusObject.GetProperty(ModemSMSInterface + ".State")
	if err != nil {
		return nil, err
	}
	sms.State = SMSState(variant.Value().(uint32))

	variant, err = dbusObject.GetProperty(ModemSMSInterface + ".Number")
	if err != nil {
		return nil, err
	}
	sms.Number = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSMSInterface + ".Text")
	if err != nil {
		return nil, err
	}
	sms.Text = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSMSInterface + ".Timestamp")
	if err != nil {
		return nil, err
	}
	sms.Timestamp, err = time.Parse("2006-01-02T15:04:05Z07", variant.Value().(string))
	if err != nil {
		return nil, err
	}
	return sms, nil
}

func (m *Modem) SendSMS(to string, text string) (*SMS, error) {
	path, err := m.CreateMessage(to, text)
	if err != nil {
		return nil, err
	}
	dbusObject, err := m.privateDbusObject(path)
	if err != nil {
		return nil, err
	}
	if err := dbusObject.Call(ModemSMSInterface+".Send", 0).Err; err != nil {
		return nil, err
	}
	return m.RetrieveSMS(path)
}
