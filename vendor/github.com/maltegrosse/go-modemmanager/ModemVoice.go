package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties of ModemVoice
const (
	ModemVoiceInterface  = ModemInterface + ".Voice"
	ModemVoiceObjectPath = ModemManagerObjectPath + "/Voice"
	/* Methods */
	ModemVoiceListCalls  = ModemVoiceInterface + ".ListCalls"
	ModemVoiceDeleteCall = ModemVoiceInterface + ".DeleteCall"
	ModemVoiceCreateCall = ModemVoiceInterface + ".CreateCall"

	ModemVoiceHoldAndAccept    = ModemVoiceInterface + ".HoldAndAccept"
	ModemVoiceHangupAndAccept  = ModemVoiceInterface + ".HangupAndAccept"
	ModemVoiceHangupAll        = ModemVoiceInterface + ".HangupAll"
	ModemVoiceTransfer         = ModemVoiceInterface + ".Transfer"
	ModemVoiceCallWaitingSetup = ModemVoiceInterface + ".CallWaitingSetup"
	ModemVoiceCallWaitingQuery = ModemVoiceInterface + ".CallWaitingQuery"

	/* Property */
	ModemVoicePropertyCalls         = ModemVoiceInterface + ".Calls"
	ModemVoicePropertyEmergencyOnly = ModemVoiceInterface + ".EmergencyOnly"

	/* Signal */
	ModemVoiceSignalCallAdded   = "CallAdded"
	ModemVoiceSignalCallDeleted = "CallDeleted"
)

// ModemVoice interface handles Calls.
// This interface will only be available once the modem is ready to be registered in the cellular network.
// 3GPP devices will require a valid unlocked SIM card before any of the features in the interface can be used.
type ModemVoice interface {
	/* METHODS */
	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// Retrieve all Calls
	// This method should only be used once and subsequent information retrieved either by listening for
	// the org.freedesktop.ModemManager1.Modem.Voice::Added signal, or by querying the specific Call object of interest.
	ListCalls() ([]Call, error)

	// Delete a Call from the list of calls.
	// The call will be hangup if it is still active.
	DeleteCall(Call) error

	// Creates a new call object for a new outgoing call.
	// The 'Number' is the only expected property to set by the user.
	CreateCall(number string, optionalParameters ...Pair) (Call, error)

	// Place all active calls on hold, if any, and accept the next call.
	// Waiting calls have preference over held calls, so the next call being active will be any waiting call, or otherwise, any held call.
	// The user should monitor the state of all available ongoing calls to be reported of which one becomes active.
	// No error is returned if there are no waiting or held calls.
	HoldAndAccept() error

	// Hangup all active calls, if any, and accept the next call.
	// Waiting calls have preference over held calls, so the next call being active will be any waiting call, or otherwise, any held call.
	// The user should monitor the state of all available ongoing calls to be reported of which one becomes active.
	// No error is returned if there are no waiting or held calls. In this case, this method would be equivalent to calling Hangup() on the active call.
	HangupAndAccept() error

	// Hangup all active calls.
	// Depending on how the device implements the action, calls on hold or in waiting state may also be terminated.
	// No error is returned if there are no ongoing calls.
	HangupAll() error

	// Join the currently active and held calls together into a single multiparty call, but disconnects from them.
	// The affected calls will be considered terminated from the point of view of the subscriber.
	Transfer() error

	// Activates or deactivates the call waiting network service, as per 3GPP TS 22.083.
	// This operation requires communication with the network in order to complete, so the modem must be successfully registered.
	CallWaitingSetup(enable bool) error

	// Queries the status of the call waiting network service, as per 3GPP TS 22.083.
	// This operation requires communication with the network in order to complete, so the modem must be successfully registered.
	CallWaitingQuery(status bool) error

	MarshalJSON() ([]byte, error)

	/* PROPERTIES */

	// The list of calls object paths.
	GetCalls() ([]Call, error)

	// A flag indicating whether emergency calls are the only allowed ones.
	// If this flag is set, users should only attempt voice calls to emergency numbers, as standard voice calls will likely fail.
	GetEmergencyOnly() (bool, error)

	/* SIGNALS */

	// CallAdded (o path);
	// Emitted when a call has been added.
	//		o path:Object path of the new call.
	SubscribeCallAdded() <-chan *dbus.Signal
	// CallDeleted (o path);
	// Emitted when a call has been deleted.
	// 		o path:Object path of the now deleted Call.
	SubscribeCallDeleted() <-chan *dbus.Signal

	ParseCallAdded(v *dbus.Signal) (Call, error)

	Unsubscribe()
}

// NewModemVoice returns new ModemVoice Interface
func NewModemVoice(objectPath dbus.ObjectPath, modem modem) (ModemVoice, error) {
	var vo modemVoice
	vo.modem = modem
	return &vo, vo.init(ModemManagerInterface, objectPath)
}

type modemVoice struct {
	modem modem
	dbusBase
	sigChan chan *dbus.Signal
}

func (m modemVoice) GetObjectPath() dbus.ObjectPath {
	return m.obj.Path()
}

func (m modemVoice) ListCalls() (calls []Call, err error) {
	var callPaths []dbus.ObjectPath
	err = m.callWithReturn(&callPaths, ModemVoiceListCalls)
	if err != nil {
		return
	}

	for idx := range callPaths {
		singleCall, err := NewCall(callPaths[idx])
		if err != nil {
			return nil, err
		}
		calls = append(calls, singleCall)
	}
	return
}

func (m modemVoice) DeleteCall(c Call) error {
	objPath := c.GetObjectPath()
	return m.call(ModemVoiceDeleteCall, &objPath)
}

func (m modemVoice) CreateCall(number string, optionalParameters ...Pair) (c Call, err error) {
	type dynMap interface{}
	myMap := make(map[string]dynMap)
	myMap["number"] = number
	for _, pair := range optionalParameters {
		myMap[fmt.Sprint(pair.GetLeft())] = fmt.Sprint(pair.GetRight())
	}
	var path dbus.ObjectPath
	err = m.callWithReturn(&path, ModemVoiceCreateCall, &myMap)
	if err != nil {
		return nil, err
	}
	singleCall, err := NewCall(path)
	if err != nil {
		return nil, err
	}
	return singleCall, nil
}

func (m modemVoice) HoldAndAccept() error {
	return m.call(ModemVoiceHoldAndAccept)
}

func (m modemVoice) HangupAndAccept() error {
	return m.call(ModemVoiceHangupAndAccept)
}

func (m modemVoice) HangupAll() error {
	return m.call(ModemVoiceHangupAll)
}

func (m modemVoice) Transfer() error {
	return m.call(ModemVoiceTransfer)
}

func (m modemVoice) CallWaitingSetup(enable bool) error {
	return m.call(ModemVoiceCallWaitingSetup, &enable)
}

func (m modemVoice) CallWaitingQuery(status bool) error {
	return m.call(ModemVoiceCallWaitingQuery, &status)
}

func (m modemVoice) GetCalls() (c []Call, err error) {
	callPaths, err := m.getSliceObjectProperty(ModemVoicePropertyCalls)
	if err != nil {
		return nil, err
	}
	for idx := range callPaths {
		singleCall, err := NewCall(callPaths[idx])
		if err != nil {
			return nil, err
		}
		c = append(c, singleCall)
	}
	return
}

func (m modemVoice) GetEmergencyOnly() (bool, error) {
	return m.getBoolProperty(ModemVoicePropertyEmergencyOnly)
}

func (m modemVoice) SubscribeCallAdded() <-chan *dbus.Signal {
	if m.sigChan != nil {
		return m.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemVoiceSignalCallAdded, fmt.Sprint(m.modem.GetObjectPath()))
	m.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	m.sigChan = make(chan *dbus.Signal, 10)
	m.conn.Signal(m.sigChan)
	return m.sigChan
}
func (m modemVoice) SubscribeCallDeleted() <-chan *dbus.Signal {
	if m.sigChan != nil {
		return m.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemVoiceSignalCallDeleted, fmt.Sprint(m.modem.GetObjectPath()))
	m.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	m.sigChan = make(chan *dbus.Signal, 10)
	m.conn.Signal(m.sigChan)
	return m.sigChan
}

func (m modemVoice) ParseCallAdded(v *dbus.Signal) (call Call, err error) {

	if strings.Contains(v.Name, ModemVoiceSignalCallAdded) == false {
		return nil, errors.New("error by parsing calladded signal")
	}

	if len(v.Body) != 1 {
		err = errors.New("error by parsing activation changed signal")
		return
	}
	path, ok := v.Body[0].(dbus.ObjectPath)
	if !ok {
		err = errors.New("error by parsing object path")
		return
	}

	return NewCall(path)
}

func (m modemVoice) Unsubscribe() {
	m.conn.RemoveSignal(m.sigChan)
	m.sigChan = nil
}

func (m modemVoice) MarshalJSON() ([]byte, error) {
	var callsJson [][]byte
	calls, err := m.GetCalls()
	if err != nil {
		return nil, err
	}
	for _, x := range calls {
		tmpB, err := x.MarshalJSON()
		if err != nil {
			return nil, err
		}
		callsJson = append(callsJson, tmpB)
	}
	emergencyOnly, err := m.GetEmergencyOnly()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"Calls":         callsJson,
		"EmergencyOnly": emergencyOnly,
	})
}
