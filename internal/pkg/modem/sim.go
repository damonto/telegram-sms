package modem

import (
	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
)

const (
	ModemPropertySimSlots       = modemmanager.ModemInterface + ".SimSlots"
	ModemPropertyPrimarySimSlot = modemmanager.ModemInterface + ".PrimarySimSlot"
	ModemSetPrimarySimSlot      = modemmanager.ModemInterface + ".SetPrimarySimSlot"
	ModemPropertySimActive      = modemmanager.SimInterface + ".Active"
)

func (m *Modem) GetSimSlots() ([]modemmanager.Sim, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), ModemPropertySimSlots)
	if err != nil {
		return nil, err
	}
	var simSlots = make([]modemmanager.Sim, 0, len(prop.Value().([]dbus.ObjectPath)))
	for _, obj := range prop.Value().([]dbus.ObjectPath) {
		if obj == "/" {
			continue
		}
		sim, err := modemmanager.NewSim(obj)
		if err != nil {
			return nil, err
		}
		simSlots = append(simSlots, sim)
	}
	return simSlots, nil
}

func (m *Modem) GetSimActiveStatus(objectPath dbus.ObjectPath) (bool, error) {
	prop, err := m.getProperty(objectPath, ModemPropertySimActive)
	if err != nil {
		return false, err
	}
	return prop.Value().(bool), nil
}

func (m *Modem) GetPrimarySimSlot() (uint32, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), ModemPropertyPrimarySimSlot)
	if err != nil {
		return 0, err
	}
	slot := prop.Value().(uint32)
	if slot == 0 {
		slot = 1
	}
	return slot, nil
}

func (m *Modem) SetPrimarySimSlot(slotId uint32) error {
	primarySlot, err := m.GetPrimarySimSlot()
	if err != nil {
		return err
	}
	if primarySlot == slotId {
		return nil
	}
	return m.call(m.modem.GetObjectPath(), ModemSetPrimarySimSlot, slotId)
}
