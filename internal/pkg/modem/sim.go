package modem

import (
	"github.com/godbus/dbus/v5"
)

const ModemSimInterface = ModemManagerInterface + ".Sim"

type SIM struct {
	Path               dbus.ObjectPath
	Active             bool
	Identifier         string
	Eid                string
	Imsi               string
	OperatorIdentifier string
	OperatorName       string
}

func (m *Modem) PrimarySIM() (*SIM, error) {
	return m.SIM(m.Sim.Path)
}

func (m *Modem) SIM(path dbus.ObjectPath) (*SIM, error) {
	var variant dbus.Variant
	var err error
	s := &SIM{Path: path}
	dbusObject, err := m.privateDbusObject(path)
	if err != nil {
		return nil, err
	}

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".Active")
	if err != nil {
		return nil, err
	}
	s.Active = variant.Value().(bool)

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".SimIdentifier")
	if err != nil {
		return nil, err
	}
	s.Identifier = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".Eid")
	if err != nil {
		return nil, err
	}
	s.Eid = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".Imsi")
	if err != nil {
		return nil, err
	}
	s.Imsi = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".OperatorIdentifier")
	if err != nil {
		return nil, err
	}
	s.OperatorIdentifier = variant.Value().(string)

	variant, err = dbusObject.GetProperty(ModemSimInterface + ".OperatorName")
	if err != nil {
		return nil, err
	}
	s.OperatorName = variant.Value().(string)
	return s, nil
}
