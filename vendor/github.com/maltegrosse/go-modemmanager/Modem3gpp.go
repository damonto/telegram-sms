package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
	"time"
)

// Paths of methods and properties of Modem3gpp
const (
	Modem3gppInterface = ModemInterface + ".Modem3gpp"

	/* Methods */
	Modem3gppRegister                    = Modem3gppInterface + ".Register"
	Modem3gppScan                        = Modem3gppInterface + ".Scan"
	Modem3gppSetEpsUeModeOperation       = Modem3gppInterface + ".SetEpsUeModeOperation"
	Modem3gppSetInitialEpsBearerSettings = Modem3gppInterface + ".SetInitialEpsBearerSettings"
	/* Property */
	Modem3gppPropertyImei                 = Modem3gppInterface + ".Imei"                 // readable   s
	Modem3gppPropertyRegistrationState    = Modem3gppInterface + ".RegistrationState"    // readable   u
	Modem3gppPropertyOperatorCode         = Modem3gppInterface + ".OperatorCode"         // readable   s
	Modem3gppPropertyOperatorName         = Modem3gppInterface + ".OperatorName"         // readable   s
	Modem3gppPropertyEnabledFacilityLocks = Modem3gppInterface + ".EnabledFacilityLocks" // readable   u
	// Deprecated Modem3gppPropertySubscriptionState        = Modem3gppInterface + ".SubscriptionState "       // readable   u
	Modem3gppPropertyEpsUeModeOperation       = Modem3gppInterface + ".EpsUeModeOperation"       // readable   u
	Modem3gppPropertyPco                      = Modem3gppInterface + ".Pco"                      // readable   a(ubay)
	Modem3gppPropertyInitialEpsBearer         = Modem3gppInterface + ".InitialEpsBearer"         // readable   o
	Modem3gppPropertyInitialEpsBearerSettings = Modem3gppInterface + ".InitialEpsBearerSettings" // readable   a{sv}
)

// Modem3gpp interface provides access to specific actions that may be performed in modems with 3GPP capabilities.
// This interface will only be available once the modem is ready to be registered in the cellular network.
// 3GPP devices will require a valid unlocked SIM card before any of the features in the interface can be used.
type Modem3gpp interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// Returns the Ussd Interface
	GetUssd() (Ussd, error)

	// The operator ID (ie, "MCCMNC", like "310260") to register. An empty string can be used to register to the home network.
	Register(operatorId string) error

	// results is an array of dictionaries with each array element describing a mobile network found in the scan.
	// takes up to 1 min
	Scan() (networks []Network3Gpp, err error)

	// Request a network scan (async)
	RequestScan()

	// Get latest scan result
	GetScanResults() (NetworkScanResult, error)

	// Sets the UE mode of operation for EPS.
	SetEpsUeModeOperation(mode MMModem3gppEpsUeModeOperation) error

	// Updates the default settings to be used in the initial default EPS bearer when registering to the LTE network.
	SetInitialEpsBearerSettings(property BearerProperty) error

	/* PROPERTIES */

	// The IMEI of the device.
	GetImei() (string, error)

	// A MMModem3gppRegistrationState value specifying the mobile registration status as defined in 3GPP TS 27.007 section 10.1.19.
	GetRegistrationState() (MMModem3gppRegistrationState, error)

	// Code of the operator to which the mobile is currently registered.
	// Returned in the format "MCCMNC", where MCC is the three-digit ITU E.212 Mobile Country Code and MNC is the two- or three-digit GSM Mobile Network Code. e.g. e"31026" or "310260".
	// If the MCC and MNC are not known or the mobile is not registered to a mobile network, this property will be a zero-length (blank) string.
	GetOperatorCode() (string, error)

	// parsed from operator code
	GetMcc() (string, error)
	GetMnc() (string, error)

	// Name of the operator to which the mobile is currently registered.
	// If the operator name is not known or the mobile is not registered to a mobile network, this property will be a zero-length (blank) string.
	GetOperatorName() (string, error)

	// Bitmask of MMModem3gppFacility values for which PIN locking is enabled.
	GetEnabledFacilityLocks() ([]MMModem3gppFacility, error)

	// A MMModem3gppEpsUeModeOperation value representing the UE mode of operation for EPS, given as an unsigned integer (signature "u").
	GetEpsUeModeOperation() (MMModem3gppEpsUeModeOperation, error)

	// The raw PCOs received from the network, given as array of PCO elements (signature "a(ubay)").
	// Each PCO is defined as a sequence of 3 fields:
	//	- The session ID associated with the PCO, given as an unsigned integer value (signature "u").
	//	- The flag that indicates whether the PCO data contains the complete PCO structure received from the network, given as a boolean value (signature"b").
	//	- The raw  PCO data, given as an array of bytes (signature "ay").
	//  Currently it's only implemented for MBIM modems that support "Microsoft Basic Connect Extensions" and for the Altair LTE plugin
	GetPco() ([]RawPcoData, error)

	// The object path for the initial default EPS bearer.
	GetInitialEpsBearer() (Bearer, error)

	// List of properties requested by the device for the initial EPS bearer during LTE network attach procedure.
	// The network may decide to use different settings during the actual device attach procedure, e.g. if the device is roaming or no explicit settings were requested, so the properties shown in the org.freedesktop.ModemManager1.Modem.Modem3gpp.InitialEpsBearer:InitialEpsBearer may be totally different.
	// This is a read-only property, updating these settings should be done using the SetInitialEpsBearerSettings() method.
	GetInitialEpsBearerSettings() (property BearerProperty, err error)

	MarshalJSON() ([]byte, error)
}

// NewModem3gpp returns new Modem3gppInterface
func NewModem3gpp(objectPath dbus.ObjectPath) (Modem3gpp, error) {
	var m3gpp modem3gpp
	scanResults = NetworkScanResult{Recent: false}
	return &m3gpp, m3gpp.init(ModemManagerInterface, objectPath)
}

type modem3gpp struct {
	dbusBase
}

// NetworkScanResult represents the results of a scanned network
type NetworkScanResult struct {
	Networks     []Network3Gpp
	LastScan     time.Time
	ScanDuration float64
	Recent       bool
}

// MarshalJSON returns a byte array
func (nsr NetworkScanResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Networks":     fmt.Sprint(nsr.Networks),
		"LastScan":     nsr.LastScan,
		"ScanDuration": nsr.ScanDuration,
		"Recent":       nsr.Recent,
	})
}

func (nsr NetworkScanResult) String() string {
	return "Networks: " + fmt.Sprint(nsr.Networks) +
		", LastScan: " + fmt.Sprint(nsr.LastScan) +
		", ScanDuration: " + fmt.Sprint(nsr.ScanDuration) +
		", Recent: " + fmt.Sprint(nsr.Recent)
}

// Network3Gpp describes a mobile network found in the scan
type Network3Gpp struct {
	Status           MMModem3gppNetworkAvailability `json:"status"`         // A MMModem3gppNetworkAvailability value representing network availability status, given as an unsigned integer (signature "u"). This key will always be present.
	OperatorLong     string                         `json:"operator-long"`  // Long-format name of operator, given as a string value (signature "s"). If the name is unknown, this field should not be present.
	OperatorShort    string                         `json:"operator-short"` // Short-format name of operator, given as a string value (signature "s"). If the name is unknown, this field should not be present.
	OperatorCode     string                         `json:"operator-code"`  // Mobile code of the operator, given as a string value (signature "s"). Returned in the format "MCCMNC", where MCC is the three-digit ITU E.212 Mobile Country Code and MNC is the two- or three-digit GSM Mobile Network Code. e.g. "31026" or "310260".
	Mcc              string                         // parsed from OperatorCode
	Mnc              string                         // parsed from OperatorCode
	AccessTechnology MMModemAccessTechnology        `json:"access-technology"` // A MMModemAccessTechnology value representing the generic access technology used by this mobile network, given as an unsigned integer (signature "u").
}

// MarshalJSON returns a byte array
func (n Network3Gpp) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Status":           fmt.Sprint(n.Status),
		"OperatorLong":     n.OperatorLong,
		"OperatorShort":    n.OperatorShort,
		"OperatorCode":     n.OperatorCode,
		"Mcc":              n.Mcc,
		"Mnc":              n.Mnc,
		"AccessTechnology": fmt.Sprint(n.AccessTechnology),
	})
}

func (n Network3Gpp) String() string {
	return "Status: " + fmt.Sprint(n.Status) +
		", OperatorLong: " + n.OperatorLong +
		", OperatorShort: " + n.OperatorShort +
		", OperatorCode: " + n.OperatorCode +
		", Mcc: " + fmt.Sprint(n.Mcc) +
		", Mnc: " + fmt.Sprint(n.Mnc) +
		", AccessTechnology: " + fmt.Sprint(n.AccessTechnology)
}

type RawPcoData struct {
	SessionId uint32 // The session ID associated with the PCO, given as an unsigned integer value (signature "u").
	Complete  bool   // The flag that indicates whether the PCO data contains the complete PCO structure received from the network, given as a boolean value (signature"b").
	RawData   []byte // The raw  PCO data, given as an array of bytes (signature "ay").
}

// MarshalJSON returns a byte array
func (r RawPcoData) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"SessionId": r.SessionId,
		"Complete":  r.Complete,
		"RawData":   r.RawData,
	})
}

func (m modem3gpp) GetObjectPath() dbus.ObjectPath {
	return m.obj.Path()
}

func (m modem3gpp) GetUssd() (Ussd, error) {
	return NewUssd(m.obj.Path())
}

func (m modem3gpp) Register(operatorId string) error {
	return m.call(Modem3gppRegister, operatorId)
}

var scanResults NetworkScanResult

func (m modem3gpp) Scan() (networks []Network3Gpp, err error) {
	// takes < 1min
	start := time.Now()
	var tmpRes interface{}
	err = m.callWithReturn(&tmpRes, Modem3gppScan)
	if err != nil {
		return nil, err
	}
	scanResMap, ok := tmpRes.([]map[string]dbus.Variant)
	if ok {
		for _, el := range scanResMap {
			var network Network3Gpp
			for key, element := range el {
				switch key {
				case "status":
					tmpState, ok := element.Value().(uint32)
					if ok {
						network.Status = MMModem3gppNetworkAvailability(tmpState)
					}
				case "operator-long":
					tmpValue, ok := element.Value().(string)
					if ok {
						network.OperatorLong = tmpValue
					}
				case "operator-short":
					tmpValue, ok := element.Value().(string)
					if ok {
						network.OperatorShort = tmpValue
					}
				case "operator-code":
					tmpValue, ok := element.Value().(string)
					if ok {
						network.OperatorCode = tmpValue
						if len(tmpValue) > 4 {
							runes := []rune(tmpValue)
							subOne := string(runes[0:3])
							network.Mcc = subOne
							subTwo := string(runes[3:len(tmpValue)])
							network.Mnc = subTwo

						}
					}
				case "access-technology":
					tmp, ok := element.Value().(uint32)
					if ok {
						network.AccessTechnology = MMModemAccessTechnology(tmp)
					}
				}
			}
			networks = append(networks, network)
		}
	}
	duration := time.Since(start).Seconds()
	scanResults = NetworkScanResult{Recent: true, LastScan: time.Now(), ScanDuration: duration, Networks: networks}

	return networks, nil
}

func (m modem3gpp) RequestScan() {
	go func() {
		res, err := m.Scan()
		if err == nil {
			fmt.Println("scan done, found ", len(res), " networks.")
		} else {
			fmt.Println("error during scanning, ", err.Error())
		}

	}()
}

func (m modem3gpp) GetScanResults() (res NetworkScanResult, err error) {
	if scanResults.Recent {
		return scanResults, nil
	}
	return res, errors.New("no recent scans")

}

func (m modem3gpp) SetEpsUeModeOperation(mode MMModem3gppEpsUeModeOperation) error {
	// todo untested
	return m.call(Modem3gppSetEpsUeModeOperation, mode)
}

func (m modem3gpp) SetInitialEpsBearerSettings(property BearerProperty) error {
	// todo untested
	v := reflect.ValueOf(property)
	st := reflect.TypeOf(property)
	ignoreFields := []string{"allow-roaming", "rm-protocol", "number"}
	type dynMap interface{}
	var myMap map[string]dynMap
	myMap = make(map[string]dynMap)
	for i := 0; i < v.NumField(); i++ {
		field := st.Field(i)
		tag := field.Tag.Get("json")
		value := v.Field(i).Interface()
		if v.Field(i).IsZero() || m.Contains(ignoreFields, tag) {
			continue
		}
		myMap[tag] = value
	}
	return m.call(Modem3gppSetInitialEpsBearerSettings, &myMap)

}

func (m modem3gpp) GetImei() (string, error) {
	return m.getStringProperty(Modem3gppPropertyImei)
}

func (m modem3gpp) GetRegistrationState() (MMModem3gppRegistrationState, error) {
	res, err := m.getUint32Property(Modem3gppPropertyRegistrationState)
	if err != nil {
		return MmModem3gppRegistrationStateUnknown, err
	}
	return MMModem3gppRegistrationState(res), nil
}

func (m modem3gpp) GetOperatorCode() (string, error) {
	return m.getStringProperty(Modem3gppPropertyOperatorCode)
}

func (m modem3gpp) GetMcc() (string, error) {
	tmpValue, err := m.GetOperatorCode()
	if err != nil {
		return "", err
	}
	if len(tmpValue) > 4 {
		runes := []rune(tmpValue)
		mcc := string(runes[0:3])
		return mcc, nil

	}
	return "", nil
}

func (m modem3gpp) GetMnc() (string, error) {
	tmpValue, err := m.GetOperatorCode()
	if err != nil {
		return "", err
	}
	if len(tmpValue) > 4 {
		runes := []rune(tmpValue)
		mnc := string(runes[3:len(tmpValue)])
		return mnc, nil
	}
	return "", err
}

func (m modem3gpp) GetOperatorName() (string, error) {
	return m.getStringProperty(Modem3gppPropertyOperatorName)
}

func (m modem3gpp) GetEnabledFacilityLocks() ([]MMModem3gppFacility, error) {
	res, err := m.getUint32Property(Modem3gppPropertyEnabledFacilityLocks)
	if err != nil {
		return nil, err
	}
	var fac MMModem3gppFacility
	return fac.BitmaskToSlice(res), nil
}

func (m modem3gpp) GetEpsUeModeOperation() (MMModem3gppEpsUeModeOperation, error) {
	res, err := m.getUint32Property(Modem3gppPropertyEpsUeModeOperation)
	if err != nil {
		return MmModem3gppEpsUeModeOperationUnknown, err
	}
	return MMModem3gppEpsUeModeOperation(res), nil
}

func (m modem3gpp) GetPco() (data []RawPcoData, err error) {
	// todo untested
	tmpRes, err := m.getInterfaceProperty(Modem3gppPropertyPco)
	if err != nil {
		return nil, err
	}
	res, ok := tmpRes.([][]interface{})
	if ok {
		for _, seq := range res {
			if len(seq) == 3 {
				sessionId, ok := seq[0].(uint32)
				if ok {
					complete, ok := seq[1].(bool)
					if ok {
						rawData, ok := seq[2].([]byte)
						if ok {
							data = append(data, RawPcoData{SessionId: sessionId, Complete: complete, RawData: rawData})
						}
					}
				}
			}
		}
		return
	}

	return nil, errors.New("wrong type")
}

func (m modem3gpp) GetInitialEpsBearer() (Bearer, error) {

	path, err := m.getObjectProperty(Modem3gppPropertyInitialEpsBearer)
	if err != nil {
		return nil, err
	}
	if fmt.Sprint(path) == "/" {
		return nil, errors.New("no initial bearer")
	}
	return NewBearer(path)
}

func (m modem3gpp) GetInitialEpsBearerSettings() (property BearerProperty, err error) {
	tmpRes, err := m.getMapStringVariantProperty(Modem3gppPropertyInitialEpsBearerSettings)
	if err != nil {
		return property, err
	}
	if len(tmpRes) < 1 {
		return property, errors.New("no initial bearer settings found")
	}
	for key, element := range tmpRes {
		switch key {
		case "apn":
			tmpValue, ok := element.Value().(string)
			if ok {
				property.APN = tmpValue
			}

		case "ip-type":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				property.IPType = MMBearerIpFamily(tmpValue)
			}
		case "allowed-auth":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				property.AllowedAuth = MMBearerAllowedAuth(tmpValue)
			}
		case "user":
			tmpValue, ok := element.Value().(string)
			if ok {
				property.User = tmpValue
			}
		case "password":
			tmpValue, ok := element.Value().(string)
			if ok {
				property.Password = tmpValue
			}
		}

	}
	return property, nil
}

func (m modem3gpp) MarshalJSON() ([]byte, error) {
	imei, err := m.GetImei()
	if err != nil {
		return nil, err
	}
	operatorCode, err := m.GetOperatorCode()
	if err != nil {
		return nil, err
	}
	operatorName, err := m.GetOperatorName()
	if err != nil {
		return nil, err
	}
	enabledFacilityLocks, err := m.GetEnabledFacilityLocks()
	if err != nil {
		return nil, err
	}
	epsUeModeOperation, err := m.GetEpsUeModeOperation()
	if err != nil {
		return nil, err
	}
	pco, err := m.GetPco()
	if err != nil {
		return nil, err
	}
	initialEpsBearerJson := []byte("")
	initialEpsBearer, err := m.GetInitialEpsBearer()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		initialEpsBearerJson, err = initialEpsBearer.MarshalJSON()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	initialEpsBearerSettingsJson := []byte("")
	initialEpsBearerSettings, err := m.GetInitialEpsBearerSettings()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		initialEpsBearerSettingsJson, err = initialEpsBearerSettings.MarshalJSON()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return json.Marshal(map[string]interface{}{
		"Imei":                     imei,
		"OperatorCode":             operatorCode,
		"OperatorName":             operatorName,
		"EnabledFacilityLocks":     enabledFacilityLocks,
		"EpsUeModeOperation":       epsUeModeOperation,
		"Pco":                      pco,
		"InitialEpsBearer":         initialEpsBearerJson,
		"InitialEpsBearerSettings": initialEpsBearerSettingsJson,
	})
}
