package modemmanager

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	BearerInterface = ModemManagerInterface + ".Bearer"

	/* Methods */
	BearerConnect    = BearerInterface + ".Connect"
	BearerDisconnect = BearerInterface + ".Disconnect"

	/* Property */
	BearerPropertyInterface  = BearerInterface + ".Interface"  // readable   s
	BearerPropertyConnected  = BearerInterface + ".Connected"  // readable   b
	BearerPropertySuspended  = BearerInterface + ".Suspended"  // readable   b
	BearerPropertyIp4Config  = BearerInterface + ".Ip4Config"  // readable   a{sv}
	BearerPropertyIp6Config  = BearerInterface + ".Ip6Config"  // readable   a{sv}
	BearerPropertyStats      = BearerInterface + ".Stats"      // readable   a{sv}
	BearerPropertyIpTimeout  = BearerInterface + ".IpTimeout"  // readable   u
	BearerPropertyBearerType = BearerInterface + ".BearerType" // readable   u
	BearerPropertyProperties = BearerInterface + ".Properties" // readable   a{sv}

)

// Bearer interface provides access to specific actions that may be performed on available bearers.
type Bearer interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// Requests activation of a packet data connection with the network using this bearer's properties. Upon successful
	// activation, the modem can send and receive packet data and, depending on the addressing capability of the
	// modem, a connection manager may need to start PPP, perform DHCP, or assign the IP address returned by the
	// modem to the data interface. Upon successful return, the "Ip4Config" and/or "Ip6Config" properties become
	// valid and may contain IP configuration information for the data interface associated with this bearer.
	Connect() error

	// Disconnect and deactivate this packet data connection.
	// Any ongoing data session will be terminated and IP addresses become invalid when this method is called.
	Disconnect() error

	MarshalJSON() ([]byte, error)

	/* PROPERTIES */

	// The operating system name for the network data interface that provides packet data using this bearer.
	// Connection managers must configure this interface depending on the IP "method" given by the "Ip4Config" or
	// "Ip6Config" properties set by bearer activation.
	// If MM_BEARER_IP_METHOD_STATIC or MM_BEARER_IP_METHOD_DHCP methods are given, the interface will be an
	// ethernet-style interface suitable for DHCP or setting static IP configuration on, while if the
	// MM_BEARER_IP_METHOD_PPP method is given, the interface will be a serial TTY which must then have PPP
	// run over it.
	GetInterface() (string, error)

	// Indicates whether or not the bearer is connected and thus whether packet data communication using this bearer is possible.
	GetConnected() (bool, error)

	// In some devices, packet data service will be suspended while the device is handling other communication,
	// like a voice call. If packet data service is suspended (but not deactivated) this property will be TRUE.
	GetSuspended() (bool, error)

	// If the bearer was configured for IPv4 addressing, upon activation this property contains the
	// addressing details for assignment to the data interface.
	// Mandatory Item: method
	// If the bearer specifies configuration via PPP or DHCP, only the "method" item will be present.
	// Additional items which are only applicable when using the MM_BEARER_IP_METHOD_STATIC method are:
	// address, prefix, dns1, dns2, dns3 and gateway
	// This property may also include the following items when such information is available: mtu
	GetIp4Config() (BearerIpConfig, error)

	// If the bearer was configured for IPv6 addressing, upon activation this property contains the addressing
	// details for assignment to the data interface.
	// Mandatory Item: method
	// If the bearer specifies configuration via PPP or DHCP, often only the "method" item will be present.
	// IPv6 SLAAC should be used to retrieve correct addressing and DNS information via Router Advertisements
	// and DHCPv6. In some cases an IPv6 Link-Local "address" item will be present, which should be assigned
	// to the data port before performing SLAAC, as the mobile network may expect SLAAC setup to use this address.
	// Additional items which are usually only applicable when using the MM_BEARER_IP_METHOD_STATIC method are:
	// address, prefix, dns1, dns2, dns3 and gateway
	// This property may also include the following items when such information is available: mtu
	GetIp6Config() (BearerIpConfig, error)

	// If the modem supports it, this property will show statistics of the ongoing connection.
	// When the connection is disconnected automatically or explicitly by the user, the values in this
	// property will show the last values cached. The statistics are reset
	GetStats() (BearerStats, error)

	// Maximum time to wait for a successful IP establishment, when PPP is used.
	GetIpTimeout() (uint32, error)

	// A MMBearerType
	GetBearerType() (MMBearerType, error)

	// List of properties used when creating the bearer.
	GetProperties() (BearerProperty, error)

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

// NewBearer returns new Bearer Interface
func NewBearer(objectPath dbus.ObjectPath) (Bearer, error) {
	var be bearer
	return &be, be.init(ModemManagerInterface, objectPath)
}

type bearer struct {
	dbusBase
	sigChan chan *dbus.Signal
}

// BearerIpConfig represents all available ip configuration properties
type BearerIpConfig struct {
	Method   MMBearerIpMethod `json:"method"`    // Mandatory: A MMBearerIpMethod, given as an unsigned integer value (signature "u").
	Address  string           `json:"address"`   // 	IP address, given as a string value (signature "s").
	Prefix   uint32           `json:"prefix"`    // Numeric CIDR network prefix (ie, 24, 32, etc), given as an unsigned integer value (signature "u").
	Dns1     string           `json:"dns1"`      // 	IP address of the first DNS server, given as a string value (signature "s").
	Dns2     string           `json:"dns2"`      // 	IP address of the second DNS server, given as a string value (signature "s").
	Dns3     string           `json:"dns3"`      // IP address of the third DNS server, given as a string value (signature "s").
	Gateway  string           `json:"gateway"`   //  IP address of the default gateway, given as a string value (signature "s").
	Mtu      uint32           `json:"mtu"`       // Maximum transmission unit (MTU), given as an unsigned integer value (signature "u").
	IpFamily MMBearerIpFamily `json:"ip-family"` // The IpFamily, either ipv4 or ipv6
}

// MarshalJSON returns a byte array
func (bc BearerIpConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Method":    fmt.Sprint(bc.Method),
		"Address":   bc.Address,
		"Prefix":    bc.Prefix,
		"Dns1":      bc.Dns1,
		"Dns2":      bc.Dns2,
		"Dns3":      bc.Dns3,
		"Gateway":   bc.Gateway,
		"Mtu":       bc.Mtu,
		"IpFamily":  fmt.Sprint(bc.IpFamily)})
}

func (bc BearerIpConfig) String() string {
	return "Method: " + fmt.Sprint(bc.Method) +
		", Address: " + bc.Address +
		", Prefix: " + fmt.Sprint(bc.Prefix) +
		", Dns1: " + bc.Dns1 +
		", Dns2: " + bc.Dns2 +
		", Dns3: " + bc.Dns3 +
		", Gateway: " + bc.Gateway +
		", Mtu: " + fmt.Sprint(bc.Mtu) +
		", IpFamily: " + fmt.Sprint(bc.IpFamily)
}

// BearerProperty represents all properties of a bearer
type BearerProperty struct {
	APN          string                `json:"apn"`           // Access Point Name, given as a string value (signature "s"). Required in 3GPP.
	IPType       MMBearerIpFamily      `json:"ip-type"`       // Addressing type, given as a MMBearerIpFamily value (signature "u"). Optional in 3GPP and CDMA.
	AllowedAuth  MMBearerAllowedAuth   `json:"allowed-auth"`  // The authentication method to use, given as a MMBearerAllowedAuth value (signature "u"). Optional in 3GPP.
	User         string                `json:"user"`          // User name (if any) required by the network, given as a string value (signature "s"). Optional in 3GPP.
	Password     string                `json:"password"`      // Password (if any) required by the network, given as a string value (signature "s"). Optional in 3GPP.
	AllowRoaming bool                  `json:"allow-roaming"` // Flag to tell whether connection is allowed during roaming, given as a boolean value (signature "b"). Optional in 3GPP.
	RMProtocol   MMModemCdmaRmProtocol `json:"rm-protocol"`   // Protocol of the Rm interface, given as a MMModemCdmaRmProtocol value (signature "u"). Optional in CDMA.
	Number       string                `json:"number"`        // Telephone number to dial, given as a string value (signature "s"). Required in POTS.
}

// MarshalJSON returns a byte array
func (bp BearerProperty) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"APN":          bp.APN,
		"IPType":       fmt.Sprint(bp.IPType),
		"AllowedAuth":  fmt.Sprint(bp.AllowedAuth),
		"User":         bp.User,
		"Password":     bp.Password,
		"AllowRoaming": bp.AllowRoaming,
		"RMProtocol":   fmt.Sprint(bp.RMProtocol),
		"Number":       bp.Number,
	})
}

func (bp BearerProperty) String() string {
	return "APN: " + bp.APN +
		", IPType: " + fmt.Sprint(bp.IPType) +
		", AllowedAuth: " + fmt.Sprint(bp.AllowedAuth) +
		", User: " + bp.User +
		", Password: " + bp.Password +
		", AllowRoaming: " + fmt.Sprint(bp.AllowRoaming) +
		", RMProtocol: " + fmt.Sprint(bp.RMProtocol) +
		", Number: " + bp.Number
}

// BearerStats represents all stats according to the bearer
type BearerStats struct {
	RxBytes  uint64 `json:"rx-bytes"` // Number of bytes received without error, given as an unsigned 64-bit integer value (signature "t").
	TxBytes  uint64 `json:"tx-bytes"` // Number bytes transmitted without error, given as an unsigned 64-bit integer value (signature "t").
	Duration uint32 `json:"duration"` // Duration of the connection, in seconds, given as an unsigned integer value (signature "u").
}

// MarshalJSON returns a byte array
func (bs BearerStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"RxBytes":  bs.RxBytes,
		"TxBytes":  bs.TxBytes,
		"Duration": bs.Duration,
	})
}
func (bs BearerStats) String() string {
	return "RxBytes: " + fmt.Sprint(bs.RxBytes) +
		", TxBytes: " + fmt.Sprint(bs.TxBytes) +
		", Duration: " + fmt.Sprint(bs.Duration)
}
func (be bearer) GetObjectPath() dbus.ObjectPath {
	return be.obj.Path()
}

func (be bearer) Connect() error {
	return be.call(BearerConnect)
}

func (be bearer) Disconnect() error {
	return be.call(BearerDisconnect)
}

func (be bearer) GetInterface() (string, error) {
	return be.getStringProperty(BearerPropertyInterface)
}

func (be bearer) GetConnected() (bool, error) {
	return be.getBoolProperty(BearerPropertyConnected)
}

func (be bearer) GetSuspended() (bool, error) {
	return be.getBoolProperty(BearerPropertySuspended)
}

func (be bearer) GetIp4Config() (bi BearerIpConfig, err error) {
	tmpMap, err := be.getMapStringVariantProperty(BearerPropertyIp4Config)
	if err != nil {
		return bi, err
	}

	bi.IpFamily = MmBearerIpFamilyIpv4
	for key, element := range tmpMap {
		switch key {
		case "method":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Method = MMBearerIpMethod(tmpValue)
			}
		case "address":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Address = tmpValue
			}
		case "prefix":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Prefix = tmpValue
			}
		case "dns1":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns1 = tmpValue
			}
		case "dns2":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns2 = tmpValue
			}
		case "dns3":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns3 = tmpValue
			}
		case "gateway":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Gateway = tmpValue
			}
		case "mtu":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Mtu = tmpValue
			}

		}
	}
	return
}

func (be bearer) GetIp6Config() (bi BearerIpConfig, err error) {
	tmpMap, err := be.getMapStringVariantProperty(BearerPropertyIp6Config)
	if err != nil {
		return bi, err
	}
	bi.IpFamily = MmBearerIpFamilyIpv6
	for key, element := range tmpMap {
		switch key {
		case "method":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Method = MMBearerIpMethod(tmpValue)
			}
		case "address":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Address = tmpValue
			}
		case "prefix":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Prefix = tmpValue
			}
		case "dns1":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns1 = tmpValue
			}
		case "dns2":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns2 = tmpValue
			}
		case "dns3":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Dns3 = tmpValue
			}
		case "gateway":
			tmpValue, ok := element.Value().(string)
			if ok {
				bi.Gateway = tmpValue
			}
		case "mtu":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bi.Mtu = tmpValue
			}

		}
	}
	return
}

func (be bearer) GetStats() (br BearerStats, err error) {
	tmpMap, err := be.getMapStringVariantProperty(BearerPropertyStats)
	if err != nil {
		return br, err
	}
	for key, element := range tmpMap {
		switch key {
		case "duration":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				br.Duration = tmpValue
			}
		case "rx-bytes":
			tmpValue, ok := element.Value().(uint64)
			if ok {
				br.RxBytes = tmpValue
			}
		case "tx-bytes":
			tmpValue, ok := element.Value().(uint64)
			if ok {
				br.TxBytes = tmpValue
			}

		}
	}
	return
}

func (be bearer) GetIpTimeout() (uint32, error) {
	return be.getUint32Property(BearerPropertyIpTimeout)
}

func (be bearer) GetBearerType() (MMBearerType, error) {
	res, err := be.getUint32Property(BearerPropertyBearerType)
	if err != nil {
		return MmBearerTypeUnknown, err
	}
	return MMBearerType(res), nil
}

func (be bearer) GetProperties() (bp BearerProperty, err error) {
	tmpMap, err := be.getMapStringVariantProperty(BearerPropertyProperties)
	if err != nil {
		return bp, err
	}
	for key, element := range tmpMap {
		switch key {
		case "apn":
			tmpValue, ok := element.Value().(string)
			if ok {
				bp.APN = tmpValue
			}
		case "ip-type":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bp.IPType = MMBearerIpFamily(tmpValue)
			}
		case "allowed-auth":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bp.AllowedAuth = MMBearerAllowedAuth(tmpValue)
			}
		case "user":
			tmpValue, ok := element.Value().(string)
			if ok {
				bp.User = tmpValue
			}
		case "password":
			tmpValue, ok := element.Value().(string)
			if ok {
				bp.Password = tmpValue
			}
		case "allow-roaming":
			tmpValue, ok := element.Value().(bool)
			if ok {
				bp.AllowRoaming = tmpValue
			}
		case "rm-protocol":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				bp.RMProtocol = MMModemCdmaRmProtocol(tmpValue)
			}
		case "number":
			tmpValue, ok := element.Value().(string)
			if ok {
				bp.Number = tmpValue
			}
		}
	}
	return
}

func (be bearer) SubscribePropertiesChanged() <-chan *dbus.Signal {
	if be.sigChan != nil {
		return be.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", dbusPropertiesChanged, fmt.Sprint(be.GetObjectPath()))
	be.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	be.sigChan = make(chan *dbus.Signal, 10)
	be.conn.Signal(be.sigChan)
	return be.sigChan
}
func (be bearer) ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	return be.parsePropertiesChanged(v)
}

func (be bearer) Unsubscribe() {
	be.conn.RemoveSignal(be.sigChan)
	be.sigChan = nil
}

func (be bearer) MarshalJSON() ([]byte, error) {

	bInterface, err := be.GetInterface()
	if err != nil {
		return nil, err
	}
	connected, err := be.GetConnected()
	if err != nil {
		return nil, err
	}
	suspended, err := be.GetSuspended()
	if err != nil {
		return nil, err
	}
	ip4Config, err := be.GetIp4Config()
	if err != nil {
		return nil, err
	}
	ip4ConfigJson, err := ip4Config.MarshalJSON()
	if err != nil {
		return nil, err
	}
	ip6Config, err := be.GetIp6Config()
	if err != nil {
		return nil, err
	}
	ip6ConfigJson, err := ip6Config.MarshalJSON()
	if err != nil {
		return nil, err
	}
	stats, err := be.GetStats()
	if err != nil {
		return nil, err
	}
	statsJson, err := stats.MarshalJSON()
	if err != nil {
		return nil, err
	}
	ipTimeout, err := be.GetIpTimeout()
	if err != nil {
		return nil, err
	}
	bearerType, err := be.GetBearerType()
	if err != nil {
		return nil, err
	}
	property, err := be.GetProperties()
	if err != nil {
		return nil, err
	}
	propertyJson, err := property.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"Interface":  bInterface,
		"Connected":  connected,
		"Suspended":  suspended,
		"Ip4Config":  ip4ConfigJson,
		"Ip6Config":  ip6ConfigJson,
		"Stats":      statsJson,
		"IpTimeout":  ipTimeout,
		"BearerType": bearerType,
		"Properties": propertyJson,
	})
}
