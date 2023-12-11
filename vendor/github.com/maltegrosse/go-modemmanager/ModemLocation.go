package modemmanager

import (
	"encoding/json"
	"errors"
	"github.com/godbus/dbus/v5"
	"strings"
	"time"
)

// Paths of methods and properties
const (
	ModemLocationInterface = ModemInterface + ".Location"

	/* Methods */
	ModemLocationSetup         = ModemLocationInterface + ".Setup"
	ModemLocationGetLocation   = ModemLocationInterface + ".GetLocation"
	ModemLocationSetSuplServer = ModemLocationInterface + ".SetSuplServer"

	ModemLocationInjectAssistanceData = ModemLocationInterface + ".InjectAssistanceData"
	ModemLocationSetGpsRefreshRate    = ModemLocationInterface + ".SetGpsRefreshRate"

	/* Property */

	ModemLocationPropertyCapabilities            = ModemLocationInterface + ".Capabilities"            //  readable   u
	ModemLocationPropertySupportedAssistanceData = ModemLocationInterface + ".SupportedAssistanceData" //  readable   u
	ModemLocationPropertyEnabled                 = ModemLocationInterface + ".Enabled"                 //  readable   u
	ModemLocationPropertySignalsLocation         = ModemLocationInterface + ".SignalsLocation"         // readable   b

	ModemLocationPropertyLocation              = ModemLocationInterface + ".Location"              // readable   a{uv}
	ModemLocationPropertySuplServer            = ModemLocationInterface + ".SuplServer"            // readable   s
	ModemLocationPropertyAssistanceDataServers = ModemLocationInterface + ".AssistanceDataServers" // readable   as
	ModemLocationPropertyGpsRefreshRate        = ModemLocationInterface + ".GpsRefreshRate"        // readable   u

)

// The ModemLocation interface allows devices to provide location information to client applications.
// Not all devices can provide this information, or even if they do, they may not be able to provide it while
// a data session is active.
// This interface will only be available once the modelo is ready to be registered in the cellular network.
// 3GPP devices will require a valid unlocked SIM card before any of the features in the interface can be used
// (including GNSS module management).
type ModemLocation interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	MarshalJSON() ([]byte, error)

	// Configure the location sources to use when gathering location information. Also enable or disable location
	// information gathering. This method may require the client to authenticate itself.
	// When signals are emitted, any client application (including malicious ones!)
	// can listen for location updates unless D-Bus permissions restrict these signals frolo certain users. If further
	// security is desired, the signal_location argument can be set to FALSE to disable location updates
	// via D-Bus signals and require applications to call authenticated APIs (like GetLocation() ) to get location information.
	// The optional MM_MODEM_LOCATION_SOURCE_AGPS_MSA and MM_MODEM_LOCATION_SOURCE_AGPS_MSB allow to
	// request MSA/MSB A-GPS operation, and they must be given along with either
	// MM_MODEM_LOCATION_SOURCE_GPS_RAW or MM_MODEM_LOCATION_SOURCE_GPS_NMEA.
	//
	//Both MM_MODEM_LOCATION_SOURCE_AGPS_MSA and MM_MODEM_LOCATION_SOURCE_AGPS_MSB cannot be given at the same time,
	// and if none given, standalone GPS is assumed.
	// 		IN u sources: Bitmask of MMModemLocationSource flags, specifying which sources should get enabled or disabled. MM_MODEM_LOCATION_SOURCE_NONE will disable all location gathering.
	//		IN b signal_location: Flag to control whether the device emits signals with the new location information. This argument is ignored when disabling location information gathering.
	Setup(sources []MMModemLocationSource, signalLocation bool) error

	// Return current location information, if any. If the modelo supports multiple location types it may return more than one. See the "Location" property for more information on the dictionary returned at location.
	// This method may require the client to authenticate itself.
	GetCurrentLocation() (CurrentLocation, error)

	// Configure the SUPL server for A-GPS.
	// IN s supl: SUPL server configuration, given either as IP:PORT or as FQDN:PORT.
	SetSuplServer(supl string) error

	// Inject assistance data to the GNSS module. The data files should be downloaded using external means
	// frolo the URLs specified in the AssistanceDataServers property.
	// The user does not need to specify the assistance data type being given.
	// There is no maximulo data size limit specified, default DBus systelo bus limits apply.
	InjectAssistanceData([]byte) error

	// Set the refresh rate of the GPS information in the API. If not explicitly set, a default of 30s will be used.
	// The refresh rate can be set to 0 to disable it, so that every update reported by the modelo is published in the interface.
	// 		IN u rate: Rate, in seconds.
	SetGpsRefreshRate(rate uint32) error

	/* PROPERTIES */

	// Bitmask of MMModemLocationSource values, specifying the supported location sources.
	GetCapabilities() ([]MMModemLocationSource, error)

	// Bitmask of MMModemLocationAssistanceDataType values, specifying the supported types of assistance data.
	GetSupportedAssistanceData() ([]MMModemLocationAssistanceDataType, error)

	// Bitmask specifying which of the supported MMModemLocationSource location sources is currently enabled in the device.
	GetEnabledLocationSources() ([]MMModemLocationSource, error)

	// TRUE if location updates will be emitted via D-Bus signals, FALSE if location updates will not be emitted.
	// See the Setup() method for more information.
	GetSignalsLocation() (bool, error)

	// Dictionary of available location information when location information gathering is enabled. If the modem
	// supports multiple location types it may return more than one here.
	// Note that if the device was told not to emit updated location information when location information
	// gathering was initially enabled, this property may not return any location information for security reasons.
	GetLocation() (CurrentLocation, error)

	// SUPL server configuration for A-GPS, given either as IP:PORT or FQDN:PORT.
	GetSuplServer() (string, error)

	// URLs from where the user can download assistance data files to inject with InjectAssistanceData().
	GetAssistanceDataServers() ([]string, error)

	// Rate of refresh of the GPS information in the interface.
	GetGpsRefreshRate() (uint32, error)
}

// NewModemLocation returns new ModemLocation Interface
func NewModemLocation(objectPath dbus.ObjectPath) (ModemLocation, error) {
	var lo modemLocation
	return &lo, lo.init(ModemManagerInterface, objectPath)
}

type modemLocation struct {
	dbusBase
}

// CurrentLocation represents all available/activated locations of the modem
type CurrentLocation struct {
	ThreeGppLacCi ThreeGppLacCiLocation `json:"3gpp-lac-ci"` // Devices supporting this capability return a string in the format "MCC,MNC,LAC,CI,TAC" (without the quotes of course)
	GpsRaw        GpsRawLocation        `json:"gps-raw"`     // Devices supporting this capability return a D-Bus dictionary (signature "a{sv}") mapping well-known keys to values with defined formats.
	GpsNmea       GpsNmeaLocation       `json:"gps-nmea"`    // Devices supporting this capability return a string containing one or more NMEA sentences (D-Bus signature 's'). The manager will cache the most recent NMEA sentence of each type for a period of time not less than 30 seconds. When reporting multiple NMEA sentences, sentences shall be separated by an ASCII Carriage Return and Line Feed (<CR><LF>) sequence.
	CdmaBs        CdmaBsLocation        `json:"cdma-bs"`     // Devices supporting this capability return a D-Bus dictionary (signature "a{sv}") mapping well-known keys to values with defined formats.
}

// MarshalJSON returns a byte array
func (cl CurrentLocation) MarshalJSON() ([]byte, error) {
	threeGppLacCiJson, err := cl.ThreeGppLacCi.MarshalJSON()
	if err != nil {
		return nil, err
	}
	gpsRawJson, err := cl.GpsRaw.MarshalJSON()
	if err != nil {
		return nil, err
	}
	gpsNmeaJson, err := cl.GpsNmea.MarshalJSON()
	if err != nil {
		return nil, err
	}
	cdmaBsJson, err := cl.CdmaBs.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"ThreeGppLacCi": threeGppLacCiJson,
		"GpsRaw":        gpsRawJson,
		"GpsNmea":       gpsNmeaJson,
		"CdmaBs":        cdmaBsJson,
	})
}

func (cl CurrentLocation) String() string {
	return returnString(cl)

}

type ThreeGppLacCiLocation struct {
	Mcc string `json:"MCC"` // This is the three-digit ITU E.212 Mobile Country Code of the network provider to which the mobile is currently registered. e.g. "310".
	Mnc string `json:"MNC"` // This is the two- or three-digit GSM Mobile Network Code of the network provider to which the mobile is currently registered. e.g. "26" or "260".
	Lac string `json:"LAC"` // This is the two-byte Location Area Code of the GSM/UMTS base station with which the mobile is registered, in upper-case hexadecimal format without leading zeros, as specified in 3GPP TS 27.007. E.g. "84CD".
	Ci  string `json:"CI"`  // This is the two- or four-byte Cell Identifier with which the mobile is registered, in upper-case hexadecimal format without leading zeros, as specified in 3GPP TS 27.007. e.g. "2BAF" or "D30156".
	Tac string `json:"TAC"` // 	This is the two-byte Location Area Code of the LTE base station with which the mobile is registered, in upper-case hexadecimal format without leading zeros, as specified in 3GPP TS 27.007. E.g. "6FFE".
}

// MarshalJSON returns a byte array
func (tgp ThreeGppLacCiLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Mcc": tgp.Mcc,
		"Mnc": tgp.Mnc,
		"Lac": tgp.Lac,
		"Ci":  tgp.Ci,
		"Tac": tgp.Tac,
	})
}

func (tgp ThreeGppLacCiLocation) String() string {
	return returnString(tgp)
}

type GpsRawLocation struct {
	UtcTime   time.Time `json:"utc-time"`  // (Required) UTC time in ISO 8601 format, given as a string value (signature "s"). e.g. 203015.
	Latitude  float64   `json:"latitude"`  // (Required) Latitude in Decimal Degrees (positive numbers mean N quadrasphere, negative mean S quadrasphere), given as a double value (signature "d"). e.g. 38.889722, meaning 38d 53' 22" N.
	Longitude float64   `json:"longitude"` // (Required) Longitude in Decimal Degrees (positive numbers mean E quadrasphere, negative mean W quadrasphere), given as a double value (signature "d"). e.g. -77.008889, meaning 77d 0' 32" W.
	Altitude  float64   `json:"altitude"`  // (Optional) Altitude above sea level in meters, given as a double value (signature "d"). e.g. 33.5.
}

func (rgps GpsRawLocation) String() string {
	return returnString(rgps)

}

// MarshalJSON returns a byte array
func (rgps GpsRawLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"UtcTime":   rgps.UtcTime,
		"Latitude":  rgps.Latitude,
		"Longitude": rgps.Longitude,
		"Altitude":  rgps.Altitude,
	})
}

type GpsNmeaLocation struct {
	NmeaSentences []string `json:"nmea-sentances"` // Devices supporting this capability return a string containing one or more NMEA sentences (D-Bus signature 's'). The manager will cache the most recent NMEA sentence of each type for a period of time not less than 30 seconds. When reporting multiple NMEA sentences, sentences shall be separated by an ASCII Carriage Return and Line Feed (<CR><LF>) sequence. The manager may discard any cached sentences older than 30 seconds.  This allows clients to read the latest positioning data as soon as possible after they start, even if the device is not providing frequent location data updates.
}

// MarshalJSON returns a byte array
func (ngps GpsNmeaLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"NmeaSentences": ngps.NmeaSentences,
	})
}

func (ngps GpsNmeaLocation) String() string {
	return returnString(ngps)
}

type CdmaBsLocation struct {
	Latitude  float64 `json:"latitude"`  // (Required) Latitude in Decimal Degrees (positive numbers mean N quadrasphere, negative mean S quadrasphere), given as a double value (signature "d"). e.g. 38.889722, meaning 38d 53' 22" N.
	Longitude float64 `json:"longitude"` // (Required) Longitude in Decimal Degrees (positive numbers mean E quadrasphere, negative mean W quadrasphere), given as a double value (signature "d"). e.g. -77.008889, meaning 77d 0' 32" W.
}

// MarshalJSON returns a byte array
func (cdma CdmaBsLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Latitude":  cdma.Latitude,
		"Longitude": cdma.Longitude,
	})
}

func (cdma CdmaBsLocation) String() string {
	return returnString(cdma)

}

func (lo modemLocation) GetObjectPath() dbus.ObjectPath {
	return lo.obj.Path()
}

func (lo modemLocation) Setup(sources []MMModemLocationSource, enableSignal bool) error {

	var tmp MMModemLocationSource
	bitmask := tmp.SliceToBitmask(sources)
	return lo.call(ModemLocationSetup, &bitmask, &enableSignal)
}

func (lo modemLocation) GetCurrentLocation() (loc CurrentLocation, err error) {
	var res map[uint32]dbus.Variant
	err = lo.callWithReturn(&res, ModemLocationGetLocation)
	if err != nil {
		return
	}

	return lo.createLocation(res)
}

func (lo modemLocation) SetSuplServer(suplServer string) error {
	return lo.call(ModemLocationSetSuplServer, &suplServer)
}

func (lo modemLocation) InjectAssistanceData(data []byte) error {
	// todo: untested
	return lo.call(ModemLocationInjectAssistanceData, &data)
}

func (lo modemLocation) SetGpsRefreshRate(rate uint32) error {
	return lo.call(ModemLocationSetGpsRefreshRate, &rate)
}

func (lo modemLocation) GetCapabilities() ([]MMModemLocationSource, error) {
	res, err := lo.getUint32Property(ModemLocationPropertyCapabilities)
	if err != nil {
		return nil, err
	}
	var tmp MMModemLocationSource
	return tmp.BitmaskToSlice(res), nil
}

func (lo modemLocation) GetSupportedAssistanceData() ([]MMModemLocationAssistanceDataType, error) {
	res, err := lo.getUint32Property(ModemLocationPropertySupportedAssistanceData)
	if err != nil {
		return nil, err
	}
	var tmp MMModemLocationAssistanceDataType
	return tmp.BitmaskToSlice(res), nil
}

func (lo modemLocation) GetEnabledLocationSources() ([]MMModemLocationSource, error) {
	res, err := lo.getUint32Property(ModemLocationPropertyEnabled)
	if err != nil {
		return nil, err
	}
	var tmp MMModemLocationSource
	return tmp.BitmaskToSlice(res), nil
}

func (lo modemLocation) GetSignalsLocation() (bool, error) {
	return lo.getBoolProperty(ModemLocationPropertySignalsLocation)
}

func (lo modemLocation) GetLocation() (locs CurrentLocation, err error) {
	res, err := lo.getMapUint32VariantProperty(ModemLocationPropertyLocation)
	if err != nil {
		return
	}
	return lo.createLocation(res)
}
func (lo modemLocation) createLocation(res map[uint32]dbus.Variant) (locs CurrentLocation, err error) {
	var sourceTypeM MMModemLocationSource
	for key, element := range res {
		sType := sourceTypeM.BitmaskToSlice(key)

		if len(sType) > 0 {
			locationType := sType[0]
			switch locationType {
			case MmModemLocationSource3gppLacCi:
				tmpString, ok := element.Value().(string)
				if ok {
					res := strings.Split(tmpString, ",")
					if len(res) == 5 {
						var three ThreeGppLacCiLocation
						three.Mcc = res[0]
						three.Mnc = res[1]
						three.Lac = res[2]
						three.Ci = res[3]
						three.Tac = res[4]
						locs.ThreeGppLacCi = three
					} else {
						return locs, errors.New("string got wrong length")
					}

				}
			case MmModemLocationSourceGpsRaw:
				tmpMap, ok := element.Value().(map[string]interface{})

				var gpsRaw GpsRawLocation
				if ok {
					for k, v := range tmpMap {
						switch k {
						case "utc-time":
							tmpTime, ok := v.(string)
							if ok {
								// Parse timestamp ("06": Year, "01": Zero Month, "02": Zero Day, "15": Hour, "04": Zero Minute, "05": Zero Second)
								t, err := time.Parse("150405", tmpTime)
								if err != nil {
									return locs, err
								}
								now := time.Now().UTC()
								// workaround as date is missing
								t = t.AddDate(now.Year(), int(now.Month()), now.Day())
								gpsRaw.UtcTime = t
							}
						case "altitude":
							tmpVal, ok := v.(float64)
							if ok {
								gpsRaw.Altitude = tmpVal
							}
						case "latitude":
							tmpVal, ok := v.(float64)
							if ok {
								gpsRaw.Latitude = tmpVal
							}
						case "longitude":
							tmpVal, ok := v.(float64)
							if ok {
								gpsRaw.Longitude = tmpVal
							}
						}
					}

					locs.GpsRaw = gpsRaw
				}

			case MmModemLocationSourceGpsNmea:
				tmpString, ok := element.Value().(string)
				if ok {
					var nmea GpsNmeaLocation
					splitString := strings.Split(tmpString, "\n")
					var tmpRes []string
					for _, nmea := range splitString {
						// just in case something weird returned
						if strings.ContainsAny(nmea, "*") {
							tmpRes = append(tmpRes, strings.TrimSpace(nmea))
						}
					}
					nmea.NmeaSentences = tmpRes
					locs.GpsNmea = nmea
				}
			case MmModemLocationSourceCdmaBs:
				// todo: untested
				tmpMap, ok := element.Value().(map[string]float64)
				if ok {
					var cdma CdmaBsLocation
					for k, v := range tmpMap {
						switch k {
						case "latitude":
							cdma.Latitude = v

						case "longitude":
							cdma.Longitude = v
						}
					}
					locs.CdmaBs = cdma
				}
			}
		}

	}
	return
}

func (lo modemLocation) GetSuplServer() (string, error) {

	return lo.getStringProperty(ModemLocationPropertySuplServer)
}

func (lo modemLocation) GetAssistanceDataServers() ([]string, error) {
	return lo.getSliceStringProperty(ModemLocationPropertyAssistanceDataServers)
}

func (lo modemLocation) GetGpsRefreshRate() (uint32, error) {
	return lo.getUint32Property(ModemLocationPropertyGpsRefreshRate)
}
func (lo modemLocation) MarshalJSON() ([]byte, error) {
	capabilities, err := lo.GetCapabilities()
	if err != nil {
		return nil, err
	}
	supportedAssistanceData, err := lo.GetSupportedAssistanceData()
	if err != nil {
		return nil, err
	}
	enabledLocationSources, err := lo.GetEnabledLocationSources()
	if err != nil {
		return nil, err
	}
	signalsLocation, err := lo.GetSignalsLocation()
	if err != nil {
		return nil, err
	}
	location, err := lo.GetLocation()
	if err != nil {
		return nil, err
	}
	locationJson, err := location.MarshalJSON()
	if err != nil {
		return nil, err
	}
	suplServer, err := lo.GetSuplServer()
	if err != nil {
		return nil, err
	}
	assistanceDataServers, err := lo.GetAssistanceDataServers()
	if err != nil {
		return nil, err
	}
	gpsRefreshRate, err := lo.GetGpsRefreshRate()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"Capabilities":            capabilities,
		"SupportedAssistanceData": supportedAssistanceData,
		"EnabledLocationSources":  enabledLocationSources,
		"SignalsLocation":         signalsLocation,
		"Location":                locationJson,
		"SuplServer":              suplServer,
		"AssistanceDataServers":   assistanceDataServers,
		"GpsRefreshRate":          gpsRefreshRate,
	})
}
