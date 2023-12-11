package modemmanager

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
)

// Paths of methods and properties
const (
	ModemSimpleInterface = ModemInterface + ".Simple"

	/* Methods */
	ModemSimpleConnect    = ModemSimpleInterface + ".Connect"
	ModemSimpleDisconnect = ModemSimpleInterface + ".Disconnect"
	ModemSimpleGetStatus  = ModemSimpleInterface + ".GetStatus"
	/* Property */

)

// The ModemSimple interface allows controlling and querying the status of Modems.
// This interface will only be available once the modem is ready to be registered in the
// cellular network. 3GPP devices will require a valid unlocked SIM card before any of the
// features in the interface can be used.
type ModemSimple interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// Do everything needed to connect the modem using the given properties.
	//This method will attempt to find a matching packet data bearer and activate it if necessary,
	//returning the bearer's IP details. If no matching bearer is found, a new bearer will be created and activated, but this operation may fail if no resources are available to complete this connection attempt (ie, if a conflicting bearer is already active).
	//This call may make a large number of changes to modem configuration based on properties passed in.
	//For example, given a PIN-locked, disabled GSM/UMTS modem, this call may unlock the SIM PIN, alter the
	//access technology preference, wait for network registration (or force registration to a specific provider),
	//create a new packet data bearer using the given "apn", and connect that bearer.
	Connect(properties SimpleProperties) (Bearer, error)

	// Disconnect an active packet data connection. while if "/" (ie, no object given) this method will disconnect all active packet data bearers.
	Disconnect(bearer Bearer) error

	// Get the general modem status.
	GetStatus() (SimpleStatus, error)
}

// SimpleProperties defines all available properties
type SimpleProperties struct {
	Pin            string                `json:"pin"`           // SIM-PIN unlock code, given as a string value (signature "s").
	OperatorId     string                `json:"operator-id"`   // ETSI MCC-MNC of a network to force registration with, given as a string value (signature "s").
	Apn            string                `json:"apn"`           // For GSM/UMTS and LTE devices the APN to use, given as a string value (signature "s").
	IpType         MMBearerIpFamily      `json:"ip-type"`       // For GSM/UMTS and LTE devices the IP addressing type to use, given as a MMBearerIpFamily value (signature "u").
	AllowedAuth    MMBearerAllowedAuth   `json:"allowed-auth"`  // The authentication method to use, given as a MMBearerAllowedAuth value (signature "u"). Optional in 3GPP.
	User           string                `json:"user"`          // User name (if any) required by the network, given as a string value (signature "s"). Optional in 3GPP.
	Password       string                `json:"password"`      // Password (if any) required by the network, given as a string value (signature "s"). Optional in 3GPP.
	Number         string                `json:"number"`        // For POTS devices the number to dial,, given as a string value (signature "s").
	AllowedRoaming bool                  `json:"allow-roaming"` // FALSE to allow only connections to home networks, given as a boolean value (signature "b").
	RmProtocol     MMModemCdmaRmProtocol `json:"rm-protocol"`   // For CDMA devices, the protocol of the Rm interface, given as a MMModemCdmaRmProtocol value (signature "u").

}

// MarshalJSON returns a byte array
func (sp SimpleProperties) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Pin":            sp.Pin,
		"OperatorId":     sp.OperatorId,
		"Apn":            sp.Apn,
		"IpType":         fmt.Sprint(sp.IpType),
		"AllowedAuth":    fmt.Sprint(sp.AllowedAuth),
		"User":           sp.User,
		"Password":       sp.Password,
		"Number":         sp.Number,
		"AllowedRoaming": sp.AllowedRoaming,
		"RmProtocol":     fmt.Sprint(sp.RmProtocol)})
}

func (sp SimpleProperties) String() string {
	return returnString(sp)
}

// SimpleStatus represents all properties of the current connection state
type SimpleStatus struct {
	State                       MMModemState                 `json:"state"`                          // A MMModemState value specifying the overall state of the modem, given as an unsigned integer value (signature "u")
	SignalQuality               uint32                       `json:"signal-quality"`                 // Signal quality value, given only when registered, as an unsigned integer value (signature "u").
	CurrentBands                []MMModemBand                `json:"current-bands"`                  // List of MMModemBand values, given only when registered, as a list of unsigned integer values (signature "au").
	AccessTechnology            MMModemAccessTechnology      `json:"access-technologies"`            // A MMModemAccessTechnology value, given only when registered, as an unsigned integer value (signature "u").
	M3GppRegistrationState      MMModem3gppRegistrationState `json:"m3gpp-registration-state"`       // A MMModem3gppRegistrationState value specifying the state of the registration, given only when registered in a 3GPP network, as an unsigned integer value (signature "u").
	M3GppOperatorCode           string                       `json:"m3gpp-operator-code"`            // Operator MCC-MNC, given only when registered in a 3GPP network, as a string value (signature "s").
	M3GppOperatorName           string                       `json:"m3gpp-operator-name"`            // Operator name, given only when registered in a 3GPP network, as a string value (signature "s").
	CdmaCdma1xRegistrationState MMModemCdmaRegistrationState `json:"cdma-cdma1x-registration-state"` // A MMModemCdmaRegistrationState value specifying the state of the registration, given only when registered in a CDMA1x network, as an unsigned integer value (signature "u").
	CdmaEvdoRegistrationState   MMModemCdmaRegistrationState `json:"cdma-evdo-registration-state"`   // A MMModemCdmaRegistrationState value specifying the state of the registration, given only when registered in a EV-DO network, as an unsigned integer value (signature "u").
	CdmaSid                     uint32                       `json:"cdma-sid"`                       // The System Identifier of the serving network, if registered in a CDMA1x network and if known. Given as an unsigned integer value (signature "u").
	CdmaNid                     uint32                       `json:"cdma-nid"`                       // The Network Identifier of the serving network, if registered in a CDMA1x network and if known. Given as an unsigned integer value (signature "u").
}

func (ss SimpleStatus) String() string {
	return returnString(ss)
}

// MarshalJSON returns a byte array
func (ss SimpleStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"State":                       fmt.Sprint(ss.State),
		"SignalQuality":               ss.SignalQuality,
		"CurrentBands":                fmt.Sprint(ss.CurrentBands),
		"AccessTechnology":            fmt.Sprint(ss.AccessTechnology),
		"M3GppRegistrationState":      fmt.Sprint(ss.M3GppRegistrationState),
		"M3GppOperatorCode":           ss.M3GppOperatorCode,
		"M3GppOperatorName":           ss.M3GppOperatorName,
		"CdmaCdma1xRegistrationState": fmt.Sprint(ss.CdmaCdma1xRegistrationState),
		"CdmaEvdoRegistrationState":   fmt.Sprint(ss.CdmaEvdoRegistrationState),
		"CdmaSid":                     ss.CdmaSid,
		"CdmaNid":                     ss.CdmaNid,
	})
}

// NewModemSimple returns new ModemSimple Interface
func NewModemSimple(objectPath dbus.ObjectPath) (ModemSimple, error) {
	var ms modemSimple
	return &ms, ms.init(ModemManagerInterface, objectPath)
}

type modemSimple struct {
	dbusBase
}

func (ms modemSimple) Connect(properties SimpleProperties) (Bearer, error) {
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
	var path dbus.ObjectPath
	err := ms.callWithReturn(&path, ModemSimpleConnect, &myMap)
	if err != nil {
		return nil, err
	}
	return NewBearer(path)
}

func (ms modemSimple) Disconnect(bearer Bearer) error {
	return ms.call(ModemSimpleDisconnect, bearer.GetObjectPath())
}

func (ms modemSimple) GetStatus() (status SimpleStatus, err error) {
	type dynMap interface{}
	var myMap map[string]dynMap
	myMap = make(map[string]dynMap)
	err = ms.callWithReturn(&myMap, ModemSimpleGetStatus)
	if err != nil {
		return status, err
	}
	for key, element := range myMap {
		switch key {
		case "state":
			tmpState, ok := element.(uint32)
			if ok {
				status.State = MMModemState(tmpState)
			}

		case "signal-quality":
			tmpSignalPair, ok := element.([]interface{})
			if ok {
				for idx := range tmpSignalPair {
					if idx == 0 {
						status.SignalQuality, _ = tmpSignalPair[idx].(uint32)
					}
				}
			}

		case "current-bands":
			var bands []MMModemBand
			tmpBands, ok := element.([]uint32)
			if ok {
				for idx := range tmpBands {
					bands = append(bands, MMModemBand(tmpBands[idx]))
				}
				status.CurrentBands = bands
			}
		// typo in dbus docs
		case "access-technologies":
			tmpValue, ok := element.(uint32)
			if ok {
				status.AccessTechnology = MMModemAccessTechnology(tmpValue)

			}

		case "m3gpp-registration-state":
			tmpValue, ok := element.(uint32)
			if ok {
				status.M3GppRegistrationState = MMModem3gppRegistrationState(tmpValue)
			}
		case "m3gpp-operator-code":
			tmpValue, ok := element.(string)
			if ok {
				status.M3GppOperatorCode = tmpValue
			}
		case "m3gpp-operator-name":
			tmpValue, ok := element.(string)
			if ok {
				status.M3GppOperatorName = tmpValue
			}
		case "cdma-cdma1x-registration-state":
			tmpValue, ok := element.(uint32)
			if ok {
				status.CdmaCdma1xRegistrationState = MMModemCdmaRegistrationState(tmpValue)
			}
		case "cdma-evdo-registration-state":
			tmpValue, ok := element.(uint32)
			if ok {
				status.CdmaEvdoRegistrationState = MMModemCdmaRegistrationState(tmpValue)
			}
		case "cdma-sid":
			tmpValue, ok := element.(uint32)
			if ok {
				status.CdmaSid = tmpValue
			}
		case "cdma-nid":
			tmpValue, ok := element.(uint32)
			if ok {
				status.CdmaNid = tmpValue
			}
		}
	}
	return status, nil

}

func (ms modemSimple) GetObjectPath() dbus.ObjectPath {
	return ms.obj.Path()
}
