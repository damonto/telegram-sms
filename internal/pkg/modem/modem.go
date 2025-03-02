package modem

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/damonto/telegram-sms/internal/pkg/util"
	"github.com/godbus/dbus/v5"
)

const ModemInterface = ModemManagerInterface + ".Modem"

type Modem struct {
	objectPath          dbus.ObjectPath
	dbusObject          dbus.BusObject
	Manufacturer        string
	EquipmentIdentifier string
	Driver              string
	Model               string
	FirmwareRevision    string
	Number              string
	PrimaryPort         string
	Ports               []ModemPort
	SimSlots            []dbus.ObjectPath
	PrimarySimSlot      uint32
	Sim                 *SIM
	State               ModemState
}

type ModemPort struct {
	PortType ModemPortType
	Device   string
}

func (m *Modem) Enable() error {
	return m.dbusObject.Call(ModemInterface+".Enable", 0, true).Err
}

func (m *Modem) Disable() error {
	return m.dbusObject.Call(ModemInterface+".Enable", 0, false).Err
}

func (m *Modem) SetPrimarySimSlot(slot uint32) error {
	return m.dbusObject.Call(ModemInterface+".SetPrimarySimSlot", 0, slot).Err
}

func (m *Modem) SignalQuality() (percent uint32, recent bool, err error) {
	variant, err := m.dbusObject.GetProperty(ModemInterface + ".SignalQuality")
	if err != nil {
		return 0, false, err
	}
	values := variant.Value().([]any)
	return values[0].(uint32), values[1].(bool), nil
}

func (m *Modem) Restart() error {
	if m.PortType() == ModemPortTypeQmi {
		if err := m.RepowerViaQMI(); err != nil {
			return err
		}
	}
	// Some older modems require disabling and enabling the modem to take effect.
	if err := m.Disable(); err != nil {
		slog.Warn("Failed to disable modem", "error", err)
	}
	return nil
}

func (m *Modem) RepowerViaQMI() error {
	device, err := m.Port(ModemPortTypeQmi)
	if err != nil {
		return err
	}
	// If multiple SIM slots aren't supported, this property will report value 0.
	// On QMI based modems the SIM slot is 1 based.
	slot := util.If(m.PrimarySimSlot > 0, m.PrimarySimSlot, 1)
	if result, err := exec.Command("/usr/bin/qmicli", "-d", device, "-p", fmt.Sprintf("--uim-sim-power-off=%d", slot)).Output(); err != nil {
		slog.Error("Failed to power off sim", "error", err, "result", string(result))
		return err
	}
	if result, err := exec.Command("/usr/bin/qmicli", "-d", device, "-p", fmt.Sprintf("--uim-sim-power-on=%d", slot)).Output(); err != nil {
		slog.Error("Failed to power on sim", "error", err, "result", string(result))
		return err
	}
	return nil
}

func (m *Modem) PortType() ModemPortType {
	supportedPorts := []ModemPortType{ModemPortTypeQmi, ModemPortTypeMbim}
	for _, port := range m.Ports {
		for _, supportedPort := range supportedPorts {
			if port.PortType == supportedPort {
				return supportedPort
			}
		}
	}
	return ModemPortTypeUnknown
}

func (m *Modem) Port(portType ModemPortType) (string, error) {
	for _, port := range m.Ports {
		if port.PortType == portType {
			return port.Device, nil
		}
	}
	return "", errors.New("port not found")
}

func (m *Modem) SystemBusPrivate() (*dbus.Conn, error) {
	dbusConn, err := dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}
	if err := dbusConn.Auth(nil); err != nil {
		dbusConn.Close()
		return nil, err
	}
	if err := dbusConn.Hello(); err != nil {
		dbusConn.Close()
		return nil, err
	}
	return dbusConn, nil
}

func (m *Modem) privateDbusObject(objectPath dbus.ObjectPath) (dbus.BusObject, error) {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	return dbusConn.Object(ModemManagerInterface, objectPath), nil
}
