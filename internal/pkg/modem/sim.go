package modem

import "github.com/godbus/dbus/v5"

const (
	ModemPropertySimSlots       = "org.freedesktop.ModemManager1.Modem.SimSlots"
	ModemPropertyPrimarySimSlot = "org.freedesktop.ModemManager1.Modem.PrimarySimSlot"
	ModemSetPrimarySimSlot      = "org.freedesktop.ModemManager1.Modem.SetPrimarySimSlot"
)

func (m *modem) GetSimSlots() (uint32, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), ModemPropertySimSlots)
	if err != nil {
		return 0, err
	}
	return uint32(len(prop.Value().([]dbus.ObjectPath))), nil
}

func (m *modem) GetPrimarySimSlot() (uint32, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), ModemPropertyPrimarySimSlot)
	if err != nil {
		return 0, err
	}
	return prop.Value().(uint32), err
}

func (m *modem) SetPrimarySimSlot(slotId uint32) error {
	primarySlot, err := m.GetPrimarySimSlot()
	if err != nil {
		return err
	}
	if primarySlot == slotId {
		return nil
	}
	return m.call(m.modem.GetObjectPath(), ModemSetPrimarySimSlot, slotId)
}
