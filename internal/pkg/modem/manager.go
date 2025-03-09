package modem

import (
	"fmt"
	"log/slog"

	"github.com/godbus/dbus/v5"
)

const (
	ModemManagerManagedObjects = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"
	ModemManagerObjectPath     = "/org/freedesktop/ModemManager1"

	ModemManagerInterface   = "org.freedesktop.ModemManager1"
	ModemManagerScanDevices = ModemManagerInterface + ".ScanDevices"

	ModemManagerInterfacesAdded   = "org.freedesktop.DBus.ObjectManager.InterfacesAdded"
	ModemManagerInterfacesRemoved = "org.freedesktop.DBus.ObjectManager.InterfacesRemoved"
)

type Manager struct {
	dbusConn   *dbus.Conn
	dbusObject dbus.BusObject
	modems     map[dbus.ObjectPath]*Modem
}

func NewManager() (*Manager, error) {
	m := &Manager{
		modems: make(map[dbus.ObjectPath]*Modem, 16),
	}
	var err error
	m.dbusConn, err = dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	m.dbusObject = m.dbusConn.Object(ModemManagerInterface, ModemManagerObjectPath)
	return m, nil
}

func (m *Manager) ScanDevices() error {
	return m.dbusObject.Call(ModemManagerScanDevices, 0).Err
}

func (m *Manager) Modems() (map[dbus.ObjectPath]*Modem, error) {
	managedObjects := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)
	if err := m.dbusObject.Call(ModemManagerManagedObjects, 0).Store(&managedObjects); err != nil {
		return nil, err
	}
	for objectPath, data := range managedObjects {
		if _, ok := data["org.freedesktop.ModemManager1.Modem"]; !ok {
			continue
		}
		modem, err := m.createModem(objectPath, data["org.freedesktop.ModemManager1.Modem"])
		if err != nil {
			slog.Error("Failed to create modem", "error", err)
			continue
		}
		m.modems[objectPath] = modem
	}
	return m.modems, nil
}

func (m *Manager) createModem(objectPath dbus.ObjectPath, data map[string]dbus.Variant) (*Modem, error) {
	modem := &Modem{
		objectPath:          objectPath,
		dbusObject:          m.dbusConn.Object(ModemManagerInterface, objectPath),
		Manufacturer:        data["Manufacturer"].Value().(string),
		EquipmentIdentifier: data["EquipmentIdentifier"].Value().(string),
		Driver:              data["Drivers"].Value().([]string)[0],
		Model:               data["Model"].Value().(string),
		FirmwareRevision:    data["Revision"].Value().(string),
		State:               ModemState(data["State"].Value().(int32)),
		PrimaryPort:         fmt.Sprintf("/dev/%s", data["PrimaryPort"].Value().(string)),
		PrimarySimSlot:      data["PrimarySimSlot"].Value().(uint32),
	}
	var err error
	modem.Sim, err = modem.SIM(data["Sim"].Value().(dbus.ObjectPath))
	if err != nil {
		return nil, err
	}
	if numbers := data["OwnNumbers"].Value().([]string); len(numbers) > 0 {
		modem.Number = numbers[0]
	}
	for _, port := range data["Ports"].Value().([][]interface{}) {
		modem.Ports = append(modem.Ports, ModemPort{
			PortType: ModemPortType(port[1].(uint32)),
			Device:   fmt.Sprintf("/dev/%s", port[0]),
		})
	}
	for _, slot := range data["SimSlots"].Value().([]dbus.ObjectPath) {
		if slot != "/" {
			modem.SimSlots = append(modem.SimSlots, slot)
		}
	}
	return modem, nil
}

func (m *Manager) Subscribe(subscriber func(map[dbus.ObjectPath]*Modem) error) error {
	if err := m.dbusConn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfacesAdded"),
		dbus.WithMatchPathNamespace("/org/freedesktop/ModemManager1"),
	); err != nil {
		return err
	}
	if err := m.dbusConn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfacesRemoved"),
		dbus.WithMatchPathNamespace("/org/freedesktop/ModemManager1"),
	); err != nil {
		return err
	}

	sig := make(chan *dbus.Signal, 10)
	m.dbusConn.Signal(sig)
	defer m.dbusConn.RemoveSignal(sig)

	for {
		event := <-sig
		modemPath := event.Body[0].(dbus.ObjectPath)
		if event.Name == ModemManagerInterfacesAdded {
			slog.Info("New modem plugged in", "path", modemPath)
			raw := event.Body[1].(map[string]map[string]dbus.Variant)
			modem, err := m.createModem(modemPath, raw["org.freedesktop.ModemManager1.Modem"])
			if err != nil {
				slog.Error("Failed to create modem", "error", err)
				continue
			}
			if modem.State == ModemStateDisabled {
				slog.Info("Enabling modem", "path", modemPath)
				if err := modem.Enable(); err != nil {
					slog.Error("Failed to enable modem", "error", err)
					continue
				}
			}
			m.modems[modemPath] = modem
			// If user user restart the ModemManager manually, Dbus will not send the InterfacesRemoved signal
			// So we need to remove the duplicate modem manually.
			m.removeDuplicate(modem, m.modems)
		} else {
			slog.Info("Modem unplugged", "path", modemPath)
			delete(m.modems, modemPath)
		}
		if err := subscriber(m.modems); err != nil {
			slog.Error("Failed to process modem", "error", err)
		}
	}
}

func (m *Manager) removeDuplicate(modem *Modem, modems map[dbus.ObjectPath]*Modem) map[dbus.ObjectPath]*Modem {
	for path, m := range modems {
		if m.EquipmentIdentifier == modem.EquipmentIdentifier {
			delete(modems, path)
		}
	}
	return modems
}
