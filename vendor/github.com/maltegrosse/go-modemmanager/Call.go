package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	CallInterface = ModemManagerInterface + ".Call"

	/* Methods */
	CallStart           = CallInterface + ".Start"
	CallAccept          = CallInterface + ".Accept"
	CallDeflect         = CallInterface + ".Deflect"
	CallJoinMultiparty  = CallInterface + ".JoinMultiparty"
	CallLeaveMultiparty = CallInterface + ".LeaveMultiparty"
	CallHangup          = CallInterface + ".Hangup"
	CallSendDtmf        = CallInterface + ".SendDtmf"

	/* Property */
	CallPropertyState       = CallInterface + ".State"       //  readable   i
	CallPropertyStateReason = CallInterface + ".StateReason" // readable   i
	CallPropertyDirection   = CallInterface + ".Direction"   // readable   i
	CallPropertyNumber      = CallInterface + ".Number"      // readable   s
	CallPropertyMultiparty  = CallInterface + ".Multiparty"  //  readable   b
	CallPropertyAudioPort   = CallInterface + ".AudioPort"   // readable   s
	CallPropertyAudioFormat = CallInterface + ".AudioFormat" // readable   a{sv}

	/* Signal */
	CallSignalDtmfReceived = "DtmfReceived"
	CallSignalStateChanged = "StateChanged"
)

// The Call interface Defines operations and properties of a single Call.
type Call interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// If the outgoing call has not yet been started, start it
	// Applicable only if state is MM_CALL_STATE_UNKNOWN and direction is MM_CALL_DIRECTION_OUTGOING.
	Start() error

	// Accept incoming call (answer).
	// Applicable only if state is MM_CALL_STATE_RINGING_IN and direction is MM_CALL_DIRECTION_INCOMING.
	Accept() error

	// Deflect an incoming or waiting call to a new number. This call will be considered terminated once the
	// deflection is performed.
	// Applicable only if state is MM_CALL_STATE_RINGING_IN or MM_CALL_STATE_WAITING and direction is
	// MM_CALL_DIRECTION_INCOMING.
	// number: new number where the call will be deflected.
	Deflect(number string) error

	// Join the currently held call into a single multiparty call with another already active call.
	// The calls will be flagged with the 'Multiparty' property while they are part of the multiparty call.
	// Applicable only if state is MM_CALL_STATE_HELD.
	JoinMultiparty() error

	// If this call is part of an ongoing multiparty call, detach it from the multiparty call, put the multiparty
	// call on hold, and activate this one alone. This operation makes this call private again between both ends of the call.
	// Applicable only if state is MM_CALL_STATE_ACTIVE or MM_CALL_STATE_HELD and the call is a multiparty call.
	LeaveMultiparty() error

	// Hangup the active call.
	// Applicable only if state is MM_CALL_STATE_UNKNOWN.
	Hangup() error

	// Send a DTMF tone (Dual Tone Multi-Frequency) (only on supported modem).
	// Applicable only if state is MM_CALL_STATE_ACTIVE.
	//		IN s dtmf: DTMF tone identifier [0-9A-D*#].
	SendDtmf(dtmf string) error

	/* PROPERTIES */
	MarshalJSON() ([]byte, error)

	// A MMCallState value, describing the state of the call.
	GetState() (MMCallState, error)

	// A MMCallStateReason value, describing why the state is changed.
	GetStateReason() (MMCallStateReason, error)

	// A MMCallDirection value, describing the direction of the call.
	GetDirection() (MMCallDirection, error)

	// The remote phone number.
	GetNumber() (string, error)

	// Whether the call is currently part of a multiparty conference call.
	GetMultiparty() (bool, error)

	// If call audio is routed via the host, the name of the kernel device that provides the audio.
	// For example, with certain Huawei USB modems, this property might be "ttyUSB2" indicating audio is
	// available via ttyUSB2 in the format described by the AudioFormat property.
	GetAudioPort() (string, error)

	// If call audio is routed via the host, a description of the audio format supported by the audio port.
	GetAudioFormat() (audioFormat, error)

	/* SIGNALS */

	// DtmfReceived (s dtmf);
	//Emitted when a DTMF tone is received (only on supported modem)
	//	s dtmf:DTMF tone identifier [0-9A-D*#].
	SubscribeDtmfReceived() <-chan *dbus.Signal

	ParseDtmfReceived(v *dbus.Signal) (string, error)

	// StateChanged (i old,
	//              i new,
	//              u reason);
	// Emitted when call changes state
	// 		i old: Old state MMCallState
	// 		i new: New state MMCallState
	// 		u reason: A MMCallStateReason value, specifying the reason for this state change.
	SubscribeStateChanged() <-chan *dbus.Signal

	// ParseStateChanged returns the parsed dbus signal
	ParseStateChanged(v *dbus.Signal) (oldState MMCallState, newState MMCallState, reason MMCallStateReason, err error)

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

// NewCall returns new Call Interface
func NewCall(objectPath dbus.ObjectPath) (Call, error) {
	var ca call
	return &ca, ca.init(ModemManagerInterface, objectPath)
}

type call struct {
	dbusBase
	sigChan chan *dbus.Signal
}

type audioFormat struct {
	Encoding   string `json:"encoding"`   // The audio encoding format. For example, "pcm" for PCM audio.
	Resolution string `json:"resolution"` // The sampling precision and its encoding format. For example, "s16le" for signed 16-bit little-endian samples
	Rate       uint32 `json:"rate"`       // The sampling rate as an unsigned integer. For example, 8000 for 8000hz.
}

// MarshalJSON returns a byte array
func (af audioFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Encoding":   af.Encoding,
		"Resolution": af.Resolution,
		"Rate":       af.Rate,
	})
}
func (af audioFormat) String() string {
	return returnString(af)

}

func (ca call) GetObjectPath() dbus.ObjectPath {
	return ca.obj.Path()
}

func (ca call) Start() error {
	return ca.call(CallStart)
}

func (ca call) Accept() error {
	return ca.call(CallAccept)
}

func (ca call) Deflect(number string) error {
	return ca.call(CallDeflect, &number)
}

func (ca call) JoinMultiparty() error {
	return ca.call(CallJoinMultiparty)
}

func (ca call) LeaveMultiparty() error {
	return ca.call(CallLeaveMultiparty)
}

func (ca call) Hangup() error {
	return ca.call(CallHangup)
}

func (ca call) SendDtmf(dtmf string) error {
	return ca.call(CallSendDtmf, &dtmf)
}

func (ca call) GetState() (MMCallState, error) {
	res, err := ca.getInt32Property(CallPropertyState)
	if err != nil {
		return MmCallStateUnknown, err
	}
	return MMCallState(res), nil
}

func (ca call) GetStateReason() (MMCallStateReason, error) {
	res, err := ca.getInt32Property(CallPropertyStateReason)
	if err != nil {
		return MmCallStateReasonUnknown, err
	}
	return MMCallStateReason(res), nil
}

func (ca call) GetDirection() (MMCallDirection, error) {
	res, err := ca.getInt32Property(CallPropertyDirection)
	if err != nil {
		return MmCallDirectionUnknown, err
	}
	return MMCallDirection(res), nil
}

func (ca call) GetNumber() (string, error) {
	return ca.getStringProperty(CallPropertyNumber)
}

func (ca call) GetMultiparty() (bool, error) {
	return ca.getBoolProperty(CallPropertyMultiparty)
}

func (ca call) GetAudioPort() (string, error) {
	return ca.getStringProperty(CallPropertyAudioPort)
}

func (ca call) GetAudioFormat() (af audioFormat, err error) {
	tmpMap, err := ca.getMapStringVariantProperty(CallPropertyAudioFormat)
	if err != nil {
		return af, err
	}
	for key, element := range tmpMap {
		switch key {
		case "encoding":
			tmpValue, ok := element.Value().(string)
			if ok {
				af.Encoding = tmpValue
			}
		case "resolution":
			tmpValue, ok := element.Value().(string)
			if ok {
				af.Resolution = tmpValue
			}
		case "rate":
			tmpValue, ok := element.Value().(uint32)
			if ok {
				af.Rate = tmpValue
			}

		}
	}
	return
}

func (ca call) SubscribeDtmfReceived() <-chan *dbus.Signal {
	if ca.sigChan != nil {
		return ca.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", CallSignalStateChanged, fmt.Sprint(ca.GetObjectPath()))
	ca.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	ca.sigChan = make(chan *dbus.Signal, 10)
	ca.conn.Signal(ca.sigChan)
	return ca.sigChan
}

func (ca call) ParseDtmfReceived(v *dbus.Signal) (dtmf string, err error) {
	if len(v.Body) != 1 {
		err = errors.New("error by parsing dtmf received signal")
		return
	}
	dtmf, ok := v.Body[0].(string)
	if !ok {
		err = errors.New("error by parsing dtmf")
		return
	}
	return

}

func (ca call) SubscribeStateChanged() <-chan *dbus.Signal {
	if ca.sigChan != nil {
		return ca.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", CallSignalStateChanged, fmt.Sprint(ca.GetObjectPath()))
	ca.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	ca.sigChan = make(chan *dbus.Signal, 10)
	ca.conn.Signal(ca.sigChan)
	return ca.sigChan
}

func (ca call) ParseStateChanged(v *dbus.Signal) (oldState MMCallState, newState MMCallState, reason MMCallStateReason, err error) {
	if len(v.Body) != 3 {
		err = errors.New("error by parsing property changed signal")
		return
	}
	oState, ok := v.Body[0].(int32)
	if !ok {
		err = errors.New("error by parsing old state")
		return
	}
	oldState = MMCallState(oState)

	nState, ok := v.Body[1].(int32)
	if !ok {
		err = errors.New("error by parsing new state")
		return
	}
	newState = MMCallState(nState)

	tmpReason, ok := v.Body[2].(uint32)
	if !ok {
		err = errors.New("error by parsing reason")
		return
	}
	reason = MMCallStateReason(tmpReason)
	return
}

func (ca call) SubscribePropertiesChanged() <-chan *dbus.Signal {
	if ca.sigChan != nil {
		return ca.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", dbusPropertiesChanged, fmt.Sprint(ca.GetObjectPath()))
	ca.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	ca.sigChan = make(chan *dbus.Signal, 10)
	ca.conn.Signal(ca.sigChan)
	return ca.sigChan
}

func (ca call) ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	return ca.parsePropertiesChanged(v)
}

func (ca call) Unsubscribe() {
	ca.conn.RemoveSignal(ca.sigChan)
	ca.sigChan = nil
}

func (ca call) MarshalJSON() ([]byte, error) {

	state, err := ca.GetState()
	if err != nil {
		return nil, err
	}
	stateReason, err := ca.GetStateReason()
	if err != nil {
		return nil, err
	}
	direction, err := ca.GetDirection()
	if err != nil {
		return nil, err
	}
	number, err := ca.GetNumber()
	if err != nil {
		return nil, err
	}
	multiparty, err := ca.GetMultiparty()
	if err != nil {
		return nil, err
	}
	audioPort, err := ca.GetAudioPort()
	if err != nil {
		return nil, err
	}
	audioFormat, err := ca.GetAudioFormat()
	if err != nil {
		return nil, err
	}
	audioFormatJson, err := audioFormat.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"State":       state,
		"StateReason": stateReason,
		"Direction":   direction,
		"Number":      number,
		"Multiparty":  multiparty,
		"AudioPort":   audioPort,
		"AudioFormat": audioFormatJson,
	})
}
