package modemmanager

import (
	"encoding/json"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	Modem3gppUssdInterface = Modem3gppInterface + ".Ussd"

	/* Methods */
	Modem3gppUssdInitiate = Modem3gppUssdInterface + ".Initiate"
	Modem3gppUssdRespond  = Modem3gppUssdInterface + ".Respond"
	Modem3gppUssdCancel   = Modem3gppUssdInterface + ".Cancel"
	/* Property */
	Modem3gppUssdPropertyState               = Modem3gppUssdInterface + ".State"               // readable   u
	Modem3gppUssdPropertyNetworkNotification = Modem3gppUssdInterface + ".NetworkNotification" // readable   s
	Modem3gppUssdProperty                    = Modem3gppUssdInterface + ".NetworkRequest"      // readable   s
)

// The Ussd interface provides access to actions based on the USSD protocol.
// This interface will only be available once the modem is ready to be registered in the
// cellular network. 3GPP devices will require a valid unlocked SIM card before any of the features
// in the interface can be used.
// Todo: Untested, USSD via QMI not available: https://gitlab.freedesktop.org/mobile-broadband/ModemManager/-/issues/26
type Ussd interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath
	// Sends a USSD command string to the network initiating a USSD session.
	// When the request is handled by the network, the method returns the response or an appropriate error. The network may be awaiting further response from the ME after returning from this method and no new command can be initiated until this one is cancelled or ended.
	// 		IN s command: The command to start the USSD session with.
	// 		OUT s reply:The network response to the command which started the USSD session.
	Initiate(command string) (reply string, err error)

	// Respond to a USSD request that is either initiated by the mobile network, or that is awaiting
	// further input after Initiate() was called.
	// IN s response:
	// The response to network-initiated USSD command, or a response to a request for further input.
	// OUT s reply:
	// The network reply to this response to the network-initiated USSD command. The reply may require
	// further responses.
	Respond(response string) (reply string, err error)

	// Cancel an ongoing USSD session, either mobile or network initiated.
	Cancel() error

	MarshalJSON() ([]byte, error)

	/* PROPERTIES */

	// A MMModem3gppUssdSessionState value, indicating the state of any ongoing USSD session.
	GetState() (MMModem3gppUssdSessionState, error)

	// Contains any network-initiated request to which no USSD response is required.
	// When no USSD session is active, or when there is no network- initiated request, this property will be a zero-length string.
	GetNetworkNotification() (string, error)

	// Contains any pending network-initiated request for a response. Client should call Respond() with the
	// appropriate response to this request.
	// When no USSD session is active, or when there is no pending network-initiated request, this property will be
	// a zero-length string.
	GetNetworkRequest() (string, error)
}

// NewUssd returns new ModemUssd Interface
func NewUssd(objectPath dbus.ObjectPath) (Ussd, error) {
	var mu ussd
	return &mu, mu.init(ModemManagerInterface, objectPath)
}

type ussd struct {
	dbusBase
}

func (mu ussd) GetObjectPath() dbus.ObjectPath {
	return mu.obj.Path()
}

func (mu ussd) Initiate(command string) (reply string, err error) {
	err = mu.callWithReturn(&reply, Modem3gppUssdInitiate, command)
	return
}

func (mu ussd) Respond(response string) (reply string, err error) {
	err = mu.callWithReturn(&reply, Modem3gppUssdRespond, response)
	return
}

func (mu ussd) Cancel() error {
	return mu.call(Modem3gppUssdCancel)
}

func (mu ussd) GetState() (MMModem3gppUssdSessionState, error) {
	res, err := mu.getUint32Property(Modem3gppUssdPropertyState)
	if err != nil {
		return MmModem3gppUssdSessionStateUnknown, err
	}
	return MMModem3gppUssdSessionState(res), nil
}

func (mu ussd) GetNetworkNotification() (string, error) {
	return mu.getStringProperty(Modem3gppUssdPropertyNetworkNotification)
}

func (mu ussd) GetNetworkRequest() (string, error) {
	return mu.getStringProperty(Modem3gppUssdProperty)
}

func (mu ussd) MarshalJSON() ([]byte, error) {
	state, err := mu.GetState()
	if err != nil {
		return nil, err
	}
	networkNotification, err := mu.GetNetworkNotification()
	if err != nil {
		return nil, err
	}
	networkRequest, err := mu.GetNetworkRequest()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"State":               state,
		"NetworkNotification": networkNotification,
		"NetworkRequest":      networkRequest,
	})
}
