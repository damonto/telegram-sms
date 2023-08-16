package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	ModemMessagingInterface = ModemInterface + ".Messaging"

	/* Methods */
	ModemMessagingList   = ModemMessagingInterface + ".List"
	ModemMessagingDelete = ModemMessagingInterface + ".Delete"
	ModemMessagingCreate = ModemMessagingInterface + ".Create"

	/* Property */
	ModemMessagingPropertyMessages          = ModemMessagingInterface + ".Messages"
	ModemMessagingPropertySupportedStorages = ModemMessagingInterface + ".SupportedStorages"
	ModemMessagingPropertyDefaultStorage    = ModemMessagingInterface + ".DefaultStorage"

	/* Signal */
	ModemMessagingSignalAdded   = "Added"
	ModemMessagingSignalDeleted = "Deleted"
)

// The ModemMessaging interface handles sending SMS messages and notification of new incoming messages.
// This interface will only be available once the modem is ready to be registered in the cellular network.
// 3GPP devices will require a valid unlocked SIM card before any of the features in the interface can
// be used (including listing stored messages).
type ModemMessaging interface {
	/* METHODS */
	// Returns object path
	GetObjectPath() dbus.ObjectPath

	MarshalJSON() ([]byte, error)

	// Retrieve all SMS messages. This method should only be used once and subsequent information retrieved either
	// by listening for the "Added" signal, or by querying the specific SMS object of interest.
	List() ([]Sms, error)

	// Delete an SMS message.
	Delete(Sms) error

	// Creates a new message object.
	// The 'Number' and either 'Text' or 'Data' properties are mandatory, others are optional.
	// If the SMSC is not specified and one is required, the default SMSC is used.
	// Optional Parameters are given has Pairs, where left side is property name as string, and right side the value as string
	// When sending, if the text/data is larger than the limit of the technology or modem, the message will be broken into multiple parts or messages.

	CreateSms(number string, text string, optionalParameters ...Pair) (Sms, error)
	CreateMms(number string, data []byte, optionalParameters ...Pair) (Sms, error)

	/* PROPERTIES */

	// The list of SMS object paths.
	GetMessages() ([]Sms, error)

	// A list of MMSmsStorage values, specifying the storages supported by this modem for storing and receiving SMS.
	GetSupportedStorages() ([]MMSmsStorage, error)

	// A MMSmsStorage value, specifying the storage to be used when receiving or storing SMS.
	GetDefaultStorage() (MMSmsStorage, error)

	/* SIGNALS */

	// Added (o path,
	//       b received);
	// Emitted when any part of a new SMS has been received or added (but not for subsequent parts, if any).
	// For messages received from the network, not all parts may have been received and the message may not be complete.
	// Check the 'State' property to determine if the message is complete.
	// 		o path: Object path of the new SMS.
	// 		b received: TRUE if the message was received from the network, as opposed to being added locally.
	//
	SubscribeAdded() <-chan *dbus.Signal

	ParseAdded(v *dbus.Signal) (Sms, bool, error)

	// Deleted (o path);
	// Emitted when a message has been deleted.
	// 		o path: Object path of the now deleted SMS.
	SubscribeDeleted() <-chan *dbus.Signal

	Unsubscribe()
}

// NewModemMessaging returns new ModemMessagingInterface
func NewModemMessaging(objectPath dbus.ObjectPath) (ModemMessaging, error) {
	var me modemMessaging
	return &me, me.init(ModemManagerInterface, objectPath)
}

type modemMessaging struct {
	dbusBase
	sigChan chan *dbus.Signal
}

func (me modemMessaging) GetObjectPath() dbus.ObjectPath {
	return me.obj.Path()
}

func (me modemMessaging) List() (sms []Sms, err error) {
	var smsPaths []dbus.ObjectPath
	err = me.callWithReturn(&smsPaths, ModemMessagingList)
	if err != nil {
		return
	}

	for idx := range smsPaths {
		singleSms, err := NewSms(smsPaths[idx])
		if err != nil {
			return nil, err
		}
		sms = append(sms, singleSms)
	}
	return
}

func (me modemMessaging) Delete(sms Sms) error {
	objPath := sms.GetObjectPath()
	return me.call(ModemMessagingDelete, &objPath)
}

func (me modemMessaging) CreateSms(number string, text string, optionalParameters ...Pair) (Sms, error) {
	type dynMap interface{}
	var myMap map[string]dynMap
	myMap = make(map[string]dynMap)
	myMap["number"] = number
	myMap["text"] = text
	for _, pair := range optionalParameters {
		myMap[fmt.Sprint(pair.GetLeft())] = fmt.Sprint(pair.GetRight())
	}
	var path dbus.ObjectPath
	err := me.callWithReturn(&path, ModemMessagingCreate, &myMap)
	if err != nil {
		return nil, err
	}
	singleSms, err := NewSms(path)
	if err != nil {
		return nil, err
	}
	return singleSms, nil
}

func (me modemMessaging) CreateMms(number string, data []byte, optionalParameters ...Pair) (Sms, error) {
	// todo: untested
	type dynMap interface{}
	var myMap map[string]dynMap
	myMap = make(map[string]dynMap)
	myMap["number"] = number
	myMap["data"] = data
	for _, pair := range optionalParameters {
		myMap[fmt.Sprint(pair.GetLeft())] = fmt.Sprint(pair.GetRight())
	}
	var path dbus.ObjectPath
	err := me.callWithReturn(&path, ModemMessagingCreate, &myMap)
	if err != nil {
		return nil, err
	}
	singleSms, err := NewSms(path)
	if err != nil {
		return nil, err
	}
	return singleSms, nil
}

func (me modemMessaging) GetMessages() (sms []Sms, err error) {
	smsPaths, err := me.getSliceObjectProperty(ModemMessagingPropertyMessages)
	if err != nil {
		return
	}
	for idx := range smsPaths {
		singleSms, err := NewSms(smsPaths[idx])
		if err != nil {
			return nil, err
		}
		sms = append(sms, singleSms)
	}
	return
}

func (me modemMessaging) GetSupportedStorages() (storages []MMSmsStorage, err error) {
	s, err := me.getSliceUint32Property(ModemMessagingPropertySupportedStorages)
	if err != nil {
		return
	}
	for _, c := range s {
		storages = append(storages, MMSmsStorage(c))
	}
	return
}

func (me modemMessaging) GetDefaultStorage() (storage MMSmsStorage, err error) {
	s, err := me.getUint32Property(ModemMessagingPropertyDefaultStorage)
	if err != nil {
		return
	}
	return MMSmsStorage(s), nil
}

func (me modemMessaging) SubscribeAdded() <-chan *dbus.Signal {
	if me.sigChan != nil {
		return me.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemMessagingSignalAdded, fmt.Sprint(me.GetObjectPath()))
	me.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	me.sigChan = make(chan *dbus.Signal, 10)
	me.conn.Signal(me.sigChan)
	return me.sigChan
}

func (me modemMessaging) ParseAdded(v *dbus.Signal) (sms Sms, received bool, err error) {
	// todo untested
	if len(v.Body) != 2 {
		err = errors.New("error by parsing added signal")
		return
	}
	path, ok := v.Body[0].(dbus.ObjectPath)
	if !ok {
		err = errors.New("error by parsing object path")
		return
	}
	sms, err = NewSms(path)
	if err != nil {
		return
	}
	received, ok = v.Body[1].(bool)
	if !ok {
		err = errors.New("error by parsing received")
		return
	}
	return
}

func (me modemMessaging) SubscribeDeleted() <-chan *dbus.Signal {
	if me.sigChan != nil {
		return me.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", ModemMessagingSignalDeleted, fmt.Sprint(me.GetObjectPath()))
	me.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	me.sigChan = make(chan *dbus.Signal, 10)
	me.conn.Signal(me.sigChan)
	return me.sigChan
}

func (me modemMessaging) Unsubscribe() {
	me.conn.RemoveSignal(me.sigChan)
	me.sigChan = nil
}

func (me modemMessaging) MarshalJSON() ([]byte, error) {
	messages, err := me.GetMessages()
	if err != nil {
		return nil, err
	}
	var messagesJson [][]byte
	for _, x := range messages {
		tmpB, err := x.MarshalJSON()
		if err != nil {
			return nil, err
		}
		messagesJson = append(messagesJson, tmpB)
	}
	supportedStorages, err := me.GetSupportedStorages()
	if err != nil {
		return nil, err
	}
	defaultStorage, err := me.GetDefaultStorage()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"Messages":          messagesJson,
		"SupportedStorages": supportedStorages,
		"DefaultStorage":    defaultStorage,
	})

}
