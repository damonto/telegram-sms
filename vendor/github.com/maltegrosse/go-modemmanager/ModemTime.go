package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"time"
)

// Paths of methods and properties of ModemTime
const (
	ModemTimeInterface = ModemInterface + ".Time"

	/* Methods */
	ModemTimeGetNetworkTime = ModemTimeInterface + ".GetNetworkTime"

	/* Property */
	ModemTimePropertyNetworkTimezone = ModemTimeInterface + ".NetworkTimezone" //  readable   a{sv}

	/* Signal */
	ModemTimeSignalNetworkTimeChanged = "NetworkTimeChanged"
)

// ModemTime interface allows clients to receive network time and timezone updates broadcast by mobile networks.
// This interface will only be available once the modem is ready to be registered in the cellular network.
// 3GPP devices will require a valid unlocked SIM card before any of the features in the interface can be used.
type ModemTime interface {
	/* METHODS */

	// get object path
	GetObjectPath() dbus.ObjectPath

	// time, and (if available) UTC offset in ISO 8601 format. If the network time is unknown, the empty string.
	// Gets the current network time in local time.
	// This method will only work if the modem tracks, or can request, the current network time; it will not attempt to use previously-received network time updates on the host to guess the current network time.
	// 		OUT s time: If the network time is known, a string containing local date,
	GetNetworkTime() (time.Time, error)

	MarshalJSON() ([]byte, error)

	/* PROPERTIES */
	// The timezone data provided by the network.
	GetNetworkTimezone() (modemTimeZone, error)

	/* SIGNALS */
	// Sent when the network time is updated.
	//		s time: A string containing date and time in ISO 8601 format.
	SubscribeNetworkTimeChanged() <-chan *dbus.Signal

	ParseNetworkTimeChanged(v *dbus.Signal) (networkTime time.Time, err error)

	Unsubscribe()
}

// NewModemTime returns new ModemTime Interface
func NewModemTime(objectPath dbus.ObjectPath) (ModemTime, error) {
	var ti modemTime
	return &ti, ti.init(ModemManagerInterface, objectPath)
}

type modemTime struct {
	dbusBase
	sigChan chan *dbus.Signal
}

// Represents the TimeZone of the Modem
type modemTimeZone struct {
	Offset      int32 `json:"offset"`       // Offset of the timezone from UTC, in minutes (including DST, if applicable), given as a signed integer value (signature "i").
	DstOffset   int32 `json:"dst-offset"`   // Amount of offset that is due to DST (daylight saving time), given as a signed integer value (signature "i").
	LeapSeconds int32 `json:"leap-seconds"` // Number of leap seconds included in the network time, given as a signed integer value (signature "i").
}

// MarshalJSON returns a byte array
func (mtz modemTimeZone) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Offset":       mtz.Offset,
		"DstOffset ":   mtz.DstOffset,
		"LeapSeconds ": mtz.LeapSeconds,
	})
}

func (mtz modemTimeZone) String() string {
	return "Offset: " + fmt.Sprint(mtz.Offset) +
		", DstOffset: " + fmt.Sprint(mtz.DstOffset) +
		", LeapSeconds: " + fmt.Sprint(mtz.LeapSeconds)
}

func (ti modemTime) GetObjectPath() dbus.ObjectPath {
	return ti.obj.Path()
}

func (ti modemTime) GetNetworkTime() (time.Time, error) {
	var tmpTime string
	err := ti.callWithReturn(&tmpTime, ModemTimeGetNetworkTime)
	if err != nil {
		return time.Now(), err
	}
	t, err := time.Parse(time.RFC3339Nano, tmpTime)
	if err != nil {
		return time.Now(), err
	}
	return t, err
}

func (ti modemTime) GetNetworkTimezone() (mTz modemTimeZone, err error) {
	tmpMap, err := ti.getMapStringVariantProperty(ModemTimePropertyNetworkTimezone)
	if err != nil {
		return mTz, err
	}
	for key, element := range tmpMap {
		switch key {
		case "offset":
			tmpValue, ok := element.Value().(int32)
			if ok {
				mTz.Offset = tmpValue
			}
		case "dst-offset":
			tmpValue, ok := element.Value().(int32)
			if ok {
				mTz.DstOffset = tmpValue
			}
		case "leap-seconds":
			tmpValue, ok := element.Value().(int32)
			if ok {
				mTz.LeapSeconds = tmpValue
			}
		}
	}
	return
}

func (ti modemTime) SubscribeNetworkTimeChanged() <-chan *dbus.Signal {
	if ti.sigChan != nil {
		return ti.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemTimeSignalNetworkTimeChanged, fmt.Sprint(ti.GetObjectPath()))
	ti.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	ti.sigChan = make(chan *dbus.Signal, 10)
	ti.conn.Signal(ti.sigChan)
	return ti.sigChan
}

func (ti modemTime) ParseNetworkTimeChanged(v *dbus.Signal) (networkTime time.Time, err error) {
	// todo: untested
	if len(v.Body) != 1 {
		err = errors.New("error by parsing network time changed signal")
		return
	}
	tmpTime, ok := v.Body[0].(string)
	if !ok {
		err = errors.New("error by parsing time string")
		return
	}
	return time.Parse(time.RFC3339Nano, tmpTime)
}

func (ti modemTime) Unsubscribe() {
	ti.conn.RemoveSignal(ti.sigChan)
	ti.sigChan = nil
}

func (ti modemTime) MarshalJSON() ([]byte, error) {
	networkTimezone, err := ti.GetNetworkTimezone()
	if err != nil {
		return nil, err
	}
	networkTimezoneJson, err := networkTimezone.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"NetworkTimezone": networkTimezoneJson,
	})
}
