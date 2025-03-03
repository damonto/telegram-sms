package modem

type ModemState int32

const (
	ModemStateFailed        ModemState = iota - 1 // The modem is unusable.
	ModemStateUnknown                             // State unknown or not reportable.
	ModemStateInitializing                        // The modem is currently being initialized.
	ModemStateLocked                              // The modem needs to be unlocked.
	ModemStateDisabled                            // The modem is not enabled and is powered down.
	ModemStateDisabling                           // The modem is currently transitioning to the @ModemStateDisabled state.
	ModemStateEnabling                            // The modem is currently transitioning to the @ModemStateEnabled state.
	ModemStateEnabled                             // The modem is enabled and powered on but not registered with a network provider and not available for data connections.
	ModemStateSearching                           // The modem is searching for a network provider to register with.
	ModemStateRegistered                          // The modem is registered with a network provider, and data connections and messaging may be available for use.
	ModemStateDisconnecting                       // The modem is disconnecting and deactivating the last active packet data bearer. This state will not be entered if more than one packet data bearer is active and one of the active bearers is deactivated.
	ModemStateConnecting                          // The modem is activating and connecting the first packet data bearer. Subsequent bearer activations when another bearer is already active do not cause this state to be entered.
	ModemStateConnected                           // One or more packet data bearers is active and connected.
)

type ModemPortType uint32

const (
	ModemPortTypeUnknown ModemPortType = iota + 1 // Unknown.
	ModemPortTypeNet                              // Net port.
	ModemPortTypeAt                               // AT port.
	ModemPortTypeQcdm                             // QCDM port.
	ModemPortTypeGps                              // GPS port.
	ModemPortTypeQmi                              // QMI port.
	ModemPortTypeMbim                             // MBIM port.
	ModemPortTypeAudio                            // Audio port.
)

type SMSState uint32

const (
	SMSStateUnknown   SMSState = iota // State unknown or not reportable.
	SMSStateStored                    // The message has been neither received nor yet sent.
	SMSStateReceiving                 // The message is being received but is not yet complete.
	SMSStateReceived                  // The message has been completely received.
	SMSStateSending                   // The message is queued for delivery.
	SMSStateSent                      // The message was successfully sent.
)

type Modem3gppRegistrationState uint32

const (
	Modem3gppRegistrationStateIdle                    Modem3gppRegistrationState = iota // Not registered, not searching for new operator to register.
	Modem3gppRegistrationStateHome                                                      // Registered on home network.
	Modem3gppRegistrationStateSearching                                                 // Not registered, searching for new operator to register with.
	Modem3gppRegistrationStateDenied                                                    // Registration denied.
	Modem3gppRegistrationStateUnknown                                                   // Unknown registration status.
	Modem3gppRegistrationStateRoaming                                                   // Registered on a roaming network.
	Modem3gppRegistrationStateHomeSmsOnly                                               // Registered for "SMS only", home network (applicable only when on LTE).
	Modem3gppRegistrationStateRoamingSmsOnly                                            // Registered for "SMS only", roaming network (applicable only when on LTE).
	Modem3gppRegistrationStateEmergencyOnly                                             // Emergency services only.
	Modem3gppRegistrationStateHomeCsfbNotPreferred                                      // Registered for "CSFB not preferred", home network (applicable only when on LTE).
	Modem3gppRegistrationStateRoamingCsfbNotPreferred                                   // Registered for "CSFB not preferred", roaming network (applicable only when on LTE).
)

type Modem3gppUssdSessionState uint32

const (
	Modem3gppUssdSessionStateUnknown      Modem3gppUssdSessionState = iota // Unknown state.
	Modem3gppUssdSessionStateIdle                                       // No active session.
	Modem3gppUssdSessionStateActive                                     // A session is active and the mobile is waiting for a response.
	Modem3gppUssdSessionStateUserResponse                               // The network is waiting for the client's response.
)
