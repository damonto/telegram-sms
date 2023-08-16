package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
)

// Paths of methods and properties
const (
	ModemCdmaInterface = ModemInterface + ".ModemCdma"

	/* Methods */
	ModemCdmaActivate       = ModemCdmaInterface + ".Activate"
	ModemCdmaActivateManual = ModemCdmaInterface + ".ActivateManual"

	/* Property */
	ModemCdmaPropertyActivationState         = ModemCdmaInterface + ".ActivationState"         //  readable   u
	ModemCdmaPropertyMeid                    = ModemCdmaInterface + ".Meid"                    // readable   s
	ModemCdmaPropertyEsn                     = ModemCdmaInterface + ".Esn"                     // readable   s
	ModemCdmaPropertySid                     = ModemCdmaInterface + ".Sid"                     // readable   u
	ModemCdmaPropertyNid                     = ModemCdmaInterface + ".Nid"                     // readable   u
	ModemCdmaPropertyCdma1xRegistrationState = ModemCdmaInterface + ".Cdma1xRegistrationState" // readable   u
	ModemCdmaPropertyEvdoRegistrationState   = ModemCdmaInterface + ".EvdoRegistrationState"   //  readable   u

	/* Signal */
	ModemCdmaSignalActivationStateChanged = "ActivationStateChanged"
)

// ModemCdma interface provides access to specific actions that may be performed in modems with CDMA capabilities.
// This interface will only be available once the modem is ready to be registered in the cellular network.
// Mixed 3GPP+3GPP2 devices will require a valid unlocked SIM card before any of the features in the interface can be used.
type ModemCdma interface {
	/* METHODS */

	// get object path
	GetObjectPath() dbus.ObjectPath

	//Provisions the modem for use with a given carrier using the modem's Over-The-Air (OTA) activation functionality, if any.
	//Some modems will reboot after this call is made.
	//	IN s carrier_code: Name of carrier, or carrier-specific code.
	Activate(carrierCode string) error

	// Sets the modem provisioning data directly, without contacting the carrier over the air.
	// Some modems will reboot after this call is made.
	ActivateManual(property CdmaProperty) error

	/* PROPERTIES */

	// A MMModemCdmaActivationState value specifying the state of the activation in the 3GPP2 network.
	GetActivationState() (MMModemCdmaActivationState, error)

	// The modem's Mobile Equipment Identifier.
	GetMeid() (string, error)

	// The modem's Electronic Serial Number (superceded by MEID but still used by older devices).
	GetEsn() (string, error)

	// The System Identifier of the serving CDMA 1x network, if known, and if the modem is registered with a CDMA 1x network.
	// See ifast.org or the mobile broadband provider database for mappings of SIDs to network providers.
	GetSid() (uint32, error)

	// The Network Identifier of the serving CDMA 1x network, if known, and if the modem is registered with a CDMA 1x network.
	GetNid() (uint32, error)

	// A MMModemCdmaRegistrationState value specifying the CDMA 1x registration state.
	GetCdma1xRegistrationState() (MMModemCdmaRegistrationState, error)

	// A MMModemCdmaRegistratiCdmaProperty onState value specifying the EVDO registration state.
	GetEvdoRegistrationState() (MMModemCdmaRegistrationState, error)

	MarshalJSON() ([]byte, error)

	/* SIGNALS */

	// The device activation state changed.
	// 		u activation_state: Current activation state, given as a MMModemCdmaActivationState.
	// 		u activation_error: Carrier-specific error code, given as a MMCdmaActivationError.
	// 		a{sv} status_changes:Properties that have changed as a result of this activation state change, including "mdn" and "min". The dictionary may be empty if the changed properties are unknown.
	SubscribeActivationStateChanged() <-chan *dbus.Signal
	// ParsePropertiesChanged parses the dbus signal
	ParseActivationStateChanged(v *dbus.Signal) (activationState MMModemCdmaActivationState, activationError MMCdmaActivationError, changedProperties map[string]dbus.Variant, err error)

	Unsubscribe()
}

// NewModemCdma returns new ModemCdma Interface
func NewModemCdma(objectPath dbus.ObjectPath) (ModemCdma, error) {
	var mc modemCdma
	return &mc, mc.init(ModemManagerInterface, objectPath)
}

type modemCdma struct {
	dbusBase
	sigChan chan *dbus.Signal
}

// CdmaProperty describes the parameters for activating manually the modem
type CdmaProperty struct {
	Spc      string `json:"spc"`        // The Service Programming Code, given as a string of exactly 6 digit characters. Mandatory parameter.
	Sid      uint16 `json:"sid"`        // The System Identification Number, given as a 16-bit unsigned integer (signature "q"). Mandatory parameter.
	Mdn      string `json:"mdn"`        // The Mobile Directory Number, given as a string of maximum 15 characters. Mandatory parameter.
	Min      string `json:"min"`        // The Mobile Identification Number, given as a string of maximum 15 characters. Mandatory parameter.
	MnHaKey  string `json:"mn-ha-key"`  // The MN-HA key, given as a string of maximum 16 characters.
	MnAaaKey string `json:"mn-aaa-key"` // The MN-AAA key, given as a string of maximum 16 characters.
	Prl      []byte `json:"prl"`        // The Preferred Roaming List, given as an array of maximum 16384 bytes.
}

// MarshalJSON returns a byte array
func (cdma CdmaProperty) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Spc":      cdma.Spc,
		"Sid":      cdma.Sid,
		"Mdn":      cdma.Mdn,
		"Min":      cdma.Min,
		"MnHaKey":  cdma.MnHaKey,
		"MnAaaKey": cdma.MnAaaKey,
		"Prl":      cdma.Prl,
	})
}
func (cdma CdmaProperty) String() string {
	return returnString(cdma)
}

func (mc modemCdma) GetObjectPath() dbus.ObjectPath {
	return mc.obj.Path()
}
func (mc modemCdma) Activate(carrierCode string) error {
	// todo: untested
	return mc.call(ModemCdmaActivate, &carrierCode)
}

func (mc modemCdma) ActivateManual(property CdmaProperty) error {
	// todo: untested
	v := reflect.ValueOf(property)
	st := reflect.TypeOf(property)

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
	return mc.call(ModemCdmaActivateManual, &myMap)
}

func (mc modemCdma) GetActivationState() (MMModemCdmaActivationState, error) {
	res, err := mc.getUint32Property(ModemCdmaPropertyActivationState)
	if err != nil {
		return MmModemCdmaActivationStateUnknown, err
	}
	return MMModemCdmaActivationState(res), nil
}

func (mc modemCdma) GetMeid() (string, error) {
	return mc.getStringProperty(ModemCdmaPropertyMeid)
}

func (mc modemCdma) GetEsn() (string, error) {
	return mc.getStringProperty(ModemCdmaPropertyEsn)
}

func (mc modemCdma) GetSid() (uint32, error) {
	return mc.getUint32Property(ModemCdmaPropertySid)
}

func (mc modemCdma) GetNid() (uint32, error) {
	return mc.getUint32Property(ModemCdmaPropertyNid)
}

func (mc modemCdma) GetCdma1xRegistrationState() (MMModemCdmaRegistrationState, error) {
	res, err := mc.getUint32Property(ModemCdmaPropertyCdma1xRegistrationState)
	if err != nil {
		return MmModemCdmaRegistrationStateUnknown, err
	}
	return MMModemCdmaRegistrationState(res), nil
}

func (mc modemCdma) GetEvdoRegistrationState() (MMModemCdmaRegistrationState, error) {
	res, err := mc.getUint32Property(ModemCdmaPropertyEvdoRegistrationState)
	if err != nil {
		return MmModemCdmaRegistrationStateUnknown, err
	}
	return MMModemCdmaRegistrationState(res), nil
}

func (mc modemCdma) SubscribeActivationStateChanged() <-chan *dbus.Signal {
	if mc.sigChan != nil {
		return mc.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemCdmaSignalActivationStateChanged, fmt.Sprint(mc.GetObjectPath()))
	mc.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	mc.sigChan = make(chan *dbus.Signal, 10)
	mc.conn.Signal(mc.sigChan)
	return mc.sigChan
}

func (mc modemCdma) ParseActivationStateChanged(v *dbus.Signal) (activationState MMModemCdmaActivationState, activationError MMCdmaActivationError, changedProperties map[string]dbus.Variant, err error) {
	if len(v.Body) != 3 {
		err = errors.New("error by parsing activation changed signal")
		return
	}
	aState, ok := v.Body[0].(uint32)
	if !ok {
		err = errors.New("error by parsing activation state")
		return
	}
	activationState = MMModemCdmaActivationState(aState)

	eState, ok := v.Body[1].(uint32)
	if !ok {
		err = errors.New("error by parsing activation error state")
		return
	}
	activationError = MMCdmaActivationError(eState)

	changedProperties, ok = v.Body[2].(map[string]dbus.Variant)
	if !ok {
		err = errors.New("error by parsing changed")
		return
	}
	return
}

func (mc modemCdma) Unsubscribe() {
	mc.conn.RemoveSignal(mc.sigChan)
	mc.sigChan = nil
}

func (mc modemCdma) MarshalJSON() ([]byte, error) {
	activationState, err := mc.GetActivationState()
	if err != nil {
		return nil, err
	}
	meid, err := mc.GetMeid()
	if err != nil {
		return nil, err
	}
	esn, err := mc.GetEsn()
	if err != nil {
		return nil, err
	}
	sid, err := mc.GetSid()
	if err != nil {
		return nil, err
	}
	nid, err := mc.GetNid()
	if err != nil {
		return nil, err
	}
	cdma1xRegistrationState, err := mc.GetCdma1xRegistrationState()
	if err != nil {
		return nil, err
	}
	evdoRegistrationState, err := mc.GetEvdoRegistrationState()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"ActivationState":         activationState,
		"Meid":                    meid,
		"Esn":                     esn,
		"Sid":                     sid,
		"Nid":                     nid,
		"Cdma1xRegistrationState": cdma1xRegistrationState,
		"EvdoRegistrationState":   evdoRegistrationState,
	})
}
