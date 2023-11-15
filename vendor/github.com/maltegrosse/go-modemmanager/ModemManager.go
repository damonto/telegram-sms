package modemmanager

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
)

// Paths of methods and properties
const (
	ModemManagerInterface = "org.freedesktop.ModemManager1"

	ModemManagerObjectPath     = "/org/freedesktop/ModemManager1"
	modemManagerMainObjectPath = "/org/freedesktop/ModemManager/"

	/* Methods */
	ModemManagerScanDevices       = ModemManagerInterface + ".ScanDevices"
	ModemManagerSetLogging        = ModemManagerInterface + ".SetLogging"
	ModemManagerReportKernelEvent = ModemManagerInterface + ".ReportKernelEvent"
	ModemManagerInhibitDevice     = ModemManagerInterface + ".InhibitDevice"

	/* Property */
	ModemManagerPropertyVersion = ModemManagerInterface + ".Version" // readable   s

)

// The ModemManager interface allows controlling and querying the status of the ModemManager daemon.
type ModemManager interface {
	/* METHODS */

	// Start a new scan for connected modem devices.
	ScanDevices() error

	// List modem devices. renamed from ListDevices to GetModems
	GetModems() ([]Modem, error)

	// Set logging verbosity.
	SetLogging(level MMLoggingLevel) error

	// Event Properties.
	// Reports a kernel event to ModemManager.
	// This method is only available if udev is not being used to report kernel events.
	// The properties dictionary is composed of key/value string pairs. The possible keys are:
	// see EventProperty and MMKernelPropertyAction
	ReportKernelEvent(EventProperties) error

	// org.freedesktop.ModemManager1.Modem:Device property. inhibit: TRUE to inhibit the modem and FALSE to uninhibit it.
	// Inhibit or uninhibit the device.
	// When the modem is inhibited ModemManager will close all its ports and unexport it from the bus, so that users of the interface are no longer able to operate with it.
	// This operation binds the inhibition request to the existence of the caller in the DBus bus. If the caller disappears from the bus, the inhibition will automatically removed.
	// 		IN s uid: the unique ID of the physical device, given in the
	// 		IN b inhibit:
	InhibitDevice(uid string, inhibit bool) error

	// The runtime version of the ModemManager daemon.
	GetVersion() (string, error)

	MarshalJSON() ([]byte, error)

	/* SIGNALS */

	// Listen to changed properties
	// returns []interface
	// index 0 = name of the interface on which the properties are defined
	// index 1 = changed properties with new values as map[string]dbus.Variant
	// index 2 = invalidated properties: changed properties but the new values are not send with them
	SubscribePropertiesChanged() <-chan *dbus.Signal

	// ParsePropertiesChanged parses the dbus signal
	ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error)
	Unsubscribe()
}

// NewModemManager returns new ModemManager Interface
func NewModemManager() (ModemManager, error) {
	var mm modemManager
	return &mm, mm.init(ModemManagerInterface, ModemManagerObjectPath)
}

type modemManager struct {
	dbusBase
	sigChan chan *dbus.Signal
}

// EventProperties  defines the properties which should be reported to the kernel
type EventProperties struct {
	Action    MMKernelPropertyAction `json:"action"`    // The type of action, given as a string value (signature "s"). This parameter is MANDATORY.
	Name      string                 `json:"name"`      // The device name, given as a string value (signature "s"). This parameter is MANDATORY.
	Subsystem string                 `json:"subsystem"` // The device subsystem, given as a string value (signature "s"). This parameter is MANDATORY.
	Uid       string                 `json:"uid"`       // The unique ID of the physical device, given as a string value (signature "s"). This parameter is OPTIONAL, if not given the sysfs path of the physical device will be used. This parameter must be the same for all devices exposed by the same physical device.
}

// MarshalJSON returns a byte array
func (ep EventProperties) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Action":    ep.Action,
		"Name ":     ep.Name,
		"Subsystem": ep.Subsystem,
		"Uid":       ep.Uid,
	})
}

func (mm modemManager) GetModems() (modems []Modem, err error) {
	devPaths, err := mm.getManagedObjects(ModemManagerInterface, ModemManagerObjectPath)
	if err != nil {
		return nil, err
	}
	for idx := range devPaths {
		modem, err := NewModem(devPaths[idx])
		if err != nil {
			return nil, err
		}
		modems = append(modems, modem)
	}
	return
}

func (mm modemManager) ScanDevices() error {
	err := mm.call(ModemManagerScanDevices)
	return err
}

func (mm modemManager) SetLogging(level MMLoggingLevel) error {
	err := mm.call(ModemManagerSetLogging, &level)
	return err
}

func (mm modemManager) ReportKernelEvent(properties EventProperties) error {
	// todo: untested
	v := reflect.ValueOf(properties)
	st := reflect.TypeOf(properties)
	type dynMap interface{}
	var myMap map[string]dynMap
	myMap = make(map[string]dynMap)
	for i := 0; i < v.NumField(); i++ {
		field := st.Field(i)
		tag := field.Tag.Get("json")
		value := v.Field(i).Interface()
		if v.Field(i).IsZero() {
			continue
		}
		myMap[tag] = value
	}
	return mm.call(ModemManagerReportKernelEvent, &myMap)
}

func (mm modemManager) InhibitDevice(uid string, inhibit bool) error {
	// todo: untested
	err := mm.call(ModemManagerInhibitDevice, &uid, &inhibit)
	return err
}

func (mm modemManager) GetVersion() (string, error) {
	v, err := mm.getStringProperty(ModemManagerPropertyVersion)
	return v, err
}
func (mm modemManager) SubscribePropertiesChanged() <-chan *dbus.Signal {
	if mm.sigChan != nil {
		return mm.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", dbusPropertiesChanged, ModemManagerObjectPath)
	mm.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	mm.sigChan = make(chan *dbus.Signal, 10)
	mm.conn.Signal(mm.sigChan)
	return mm.sigChan
}
func (mm modemManager) ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	return mm.parsePropertiesChanged(v)
}

func (mm modemManager) Unsubscribe() {
	mm.conn.RemoveSignal(mm.sigChan)
	mm.sigChan = nil
}

func (mm modemManager) MarshalJSON() ([]byte, error) {
	version, err := mm.GetVersion()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"Version": version,
	})
}
