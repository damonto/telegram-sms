package modemmanager

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	SimInterface = ModemManagerInterface + ".Sim"

	/* Methods */
	SimSendPin     = SimInterface + ".SendPin"
	SimSendSendPuk = SimInterface + ".SendPuk"
	SimEnablePin   = SimInterface + ".EnablePin"
	SimChangePin   = SimInterface + ".ChangePin"

	/* Property */
	SimPropertySimIdentifier      = SimInterface + ".SimIdentifier"      // readable   s
	SimPropertyImsi               = SimInterface + ".Imsi"               // readable   s
	SimPropertyOperatorIdentifier = SimInterface + ".OperatorIdentifier" // readable   s
	SimPropertyOperatorName       = SimInterface + ".OperatorName"       // readable   s
	SimPropertyEmergencyNumbers   = SimInterface + ".EmergencyNumbers"   // readable   as

)

// The Sim interface handles communication with SIM, USIM, and RUIM (CDMA SIM) cards.
type Sim interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// Send the PIN to unlock the SIM card.
	SendPin(pin string) error

	// Send the PUK and a new PIN to unlock the SIM card.
	SendPuk(pin string, puk string) error

	// Enable or disable the PIN checking.
	//		IN s pin: A string containing the PIN code.
	//		IN b enabled: TRUE to enable PIN checking, FALSE otherwise.
	EnablePin(pin string, enable bool) error

	// Change the PIN code.
	// 		IN s old_pin: A string containing the current PIN code.
	// I	N s new_pin: A string containing the new PIN code.
	ChangePin(oldPin string, newPin string) error

	/* PROPERTIES */

	// The ICCID of the SIM card.
	// This may be available before the PIN has been entered depending on the device itself.
	GetSimIdentifier() (string, error)

	// The IMSI of the SIM card, if any.
	GetImsi() (string, error)

	// The OperatorIdentifier
	GetOperatorIdentifier() (string, error)

	// The name of the network operator, as given by the SIM card, if known.
	GetOperatorName() (string, error)

	// List of emergency numbers programmed in the SIM card.
	// These numbers should be treated as numbers for emergency calls in addition to 112 and 911.
	GetEmergencyNumbers() ([]string, error)

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

// NewSim returns new Sim Interface
func NewSim(objectPath dbus.ObjectPath) (Sim, error) {
	var sm sim
	return &sm, sm.init(ModemManagerInterface, objectPath)
}

type sim struct {
	dbusBase
	sigChan chan *dbus.Signal
}

func (sm sim) GetObjectPath() dbus.ObjectPath {
	return sm.obj.Path()
}
func (sm sim) SendPin(pin string) error {
	return sm.call(SimSendPin, &pin)
}

func (sm sim) SendPuk(pin string, puk string) error {
	return sm.call(SimSendSendPuk, &pin, &puk)
}

func (sm sim) EnablePin(pin string, enable bool) error {
	return sm.call(SimEnablePin, &pin, &enable)
}

func (sm sim) ChangePin(oldPin string, newPin string) error {
	return sm.call(SimChangePin, &oldPin, &newPin)
}

func (sm sim) GetSimIdentifier() (string, error) {
	return sm.getStringProperty(SimPropertySimIdentifier)
}

func (sm sim) GetImsi() (string, error) {
	return sm.getStringProperty(SimPropertyImsi)
}

func (sm sim) GetOperatorIdentifier() (string, error) {
	return sm.getStringProperty(SimPropertyOperatorIdentifier)
}

func (sm sim) GetOperatorName() (string, error) {
	return sm.getStringProperty(SimPropertyOperatorName)
}

func (sm sim) GetEmergencyNumbers() ([]string, error) {
	return sm.getSliceStringProperty(SimPropertyEmergencyNumbers)
}

func (sm sim) SubscribePropertiesChanged() <-chan *dbus.Signal {
	if sm.sigChan != nil {
		return sm.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", dbusPropertiesChanged, fmt.Sprint(sm.GetObjectPath()))
	sm.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	sm.sigChan = make(chan *dbus.Signal, 10)
	sm.conn.Signal(sm.sigChan)
	return sm.sigChan
}
func (sm sim) ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	return sm.parsePropertiesChanged(v)
}

func (sm sim) Unsubscribe() {
	sm.conn.RemoveSignal(sm.sigChan)
	sm.sigChan = nil
}

func (sm sim) MarshalJSON() ([]byte, error) {
	simIdentifier, err := sm.GetSimIdentifier()
	if err != nil {
		return nil, err
	}
	imsi, err := sm.GetImsi()
	if err != nil {
		return nil, err
	}
	operatorIdentifier, err := sm.GetOperatorIdentifier()
	if err != nil {
		return nil, err
	}
	operatorName, err := sm.GetOperatorName()
	if err != nil {
		return nil, err
	}
	emergencyNumbers, err := sm.GetEmergencyNumbers()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"SimIdentifier":      simIdentifier,
		"Imsi":               imsi,
		"OperatorIdentifier": operatorIdentifier,
		"OperatorName":       operatorName,
		"EmergencyNumbers":   emergencyNumbers,
	})
}
