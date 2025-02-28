package modem

type ModemState int32

const (
	ModemStateFailed        ModemState = -1 // The modem is unusable.
	ModemStateUnknown       ModemState = 0  // State unknown or not reportable.
	ModemStateInitializing  ModemState = 1  // The modem is currently being initialized.
	ModemStateLocked        ModemState = 2  // The modem needs to be unlocked.
	ModemStateDisabled      ModemState = 3  // The modem is not enabled and is powered down.
	ModemStateDisabling     ModemState = 4  // The modem is currently transitioning to the @ModemStateDisabled state.
	ModemStateEnabling      ModemState = 5  // The modem is currently transitioning to the @ModemStateEnabled state.
	ModemStateEnabled       ModemState = 6  // The modem is enabled and powered on but not registered with a network provider and not available for data connections.
	ModemStateSearching     ModemState = 7  // The modem is searching for a network provider to register with.
	ModemStateRegistered    ModemState = 8  // The modem is registered with a network provider, and data connections and messaging may be available for use.
	ModemStateDisconnecting ModemState = 9  // The modem is disconnecting and deactivating the last active packet data bearer. This state will not be entered if more than one packet data bearer is active and one of the active bearers is deactivated.
	ModemStateConnecting    ModemState = 10 // The modem is activating and connecting the first packet data bearer. Subsequent bearer activations when another bearer is already active do not cause this state to be entered.
	ModemStateConnected     ModemState = 11 // One or more packet data bearers is active and connected.
)

type ModemPortType uint32

const (
	ModemPortTypeUnknown ModemPortType = 1 // Unknown.
	ModemPortTypeNet     ModemPortType = 2 // Net port.
	ModemPortTypeAt      ModemPortType = 3 // AT port.
	ModemPortTypeQcdm    ModemPortType = 4 // QCDM port.
	ModemPortTypeGps     ModemPortType = 5 // GPS port.
	ModemPortTypeQmi     ModemPortType = 6 // QMI port.
	ModemPortTypeMbim    ModemPortType = 7 // MBIM port.
	ModemPortTypeAudio   ModemPortType = 8 // Audio port.
)

type SMSState uint32

const (
	SMSStateUnknown   SMSState = 0 // State unknown or not reportable.
	SMSStateStored    SMSState = 1 // The message has been neither received nor yet sent.
	SMSStateReceiving SMSState = 2 // The message is being received but is not yet complete.
	SMSStateReceived  SMSState = 3 // The message has been completely received.
	SMSStateSending   SMSState = 4 // The message is queued for delivery.
	SMSStateSent      SMSState = 5 // The message was successfully sent.
)

type Modem3gppRegistrationState uint32

const (
	Modem3gppRegistrationStateIdle                    Modem3gppRegistrationState = 0  // Not registered, not searching for new operator to register.
	Modem3gppRegistrationStateHome                    Modem3gppRegistrationState = 1  // Registered on home network.
	Modem3gppRegistrationStateSearching               Modem3gppRegistrationState = 2  // Not registered, searching for new operator to register with.
	Modem3gppRegistrationStateDenied                  Modem3gppRegistrationState = 3  // Registration denied.
	Modem3gppRegistrationStateUnknown                 Modem3gppRegistrationState = 4  // Unknown registration status.
	Modem3gppRegistrationStateRoaming                 Modem3gppRegistrationState = 5  // Registered on a roaming network.
	Modem3gppRegistrationStateHomeSmsOnly             Modem3gppRegistrationState = 6  // Registered for "SMS only", home network (applicable only when on LTE).
	Modem3gppRegistrationStateRoamingSmsOnly          Modem3gppRegistrationState = 7  // Registered for "SMS only", roaming network (applicable only when on LTE).
	Modem3gppRegistrationStateEmergencyOnly           Modem3gppRegistrationState = 8  // Emergency services only.
	Modem3gppRegistrationStateHomeCsfbNotPreferred    Modem3gppRegistrationState = 9  // Registered for "CSFB not preferred", home network (applicable only when on LTE).
	Modem3gppRegistrationStateRoamingCsfbNotPreferred Modem3gppRegistrationState = 10 // Registered for "CSFB not preferred", roaming network (applicable only when on LTE).
)

type Modem3gppUssdSessionState uint32

const (
	Modem3gppUssdSessionStateUnknown      Modem3gppUssdSessionState = 0 // Unknown state.
	Modem3gppUssdSessionStateIdle         Modem3gppUssdSessionState = 1 // No active session.
	Modem3gppUssdSessionStateActive       Modem3gppUssdSessionState = 2 // A session is active and the mobile is waiting for a response.
	Modem3gppUssdSessionStateUserResponse Modem3gppUssdSessionState = 3 // The network is waiting for the client's response.
)
