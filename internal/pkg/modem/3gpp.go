package modem

const Modem3GPPInterface = ModemInterface + ".Modem3gpp"

func (m *Modem) IMEI() (string, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".Imei")
	if err != nil {
		return "", err
	}
	return variant.Value().(string), nil
}

func (m *Modem) RegistrationState() (Modem3gppRegistrationState, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".RegistrationState")
	if err != nil {
		return 0, err
	}
	return Modem3gppRegistrationState(variant.Value().(uint32)), nil
}

func (m *Modem) OperatorCode() (string, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".OperatorCode")
	if err != nil {
		return "", err
	}
	return variant.Value().(string), nil
}

func (m *Modem) OperatorName() (string, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".OperatorName")
	if err != nil {
		return "", err
	}
	return variant.Value().(string), nil
}

func (m *Modem) InitiateUSSD(command string) (string, error) {
	var reply string
	err := m.dbusObject.Call(Modem3GPPInterface+".Ussd.Initiate", 0, command).Store(&reply)
	return reply, err
}

func (m *Modem) RespondUSSD(response string) (string, error) {
	var reply string
	err := m.dbusObject.Call(Modem3GPPInterface+".Ussd.Respond", 0, response).Store(&reply)
	return reply, err
}

func (m *Modem) CancelUSSD() error {
	return m.dbusObject.Call(Modem3GPPInterface+".Ussd.Cancel", 0).Err
}

func (m *Modem) USSDState() (Modem3gppUssdSessionState, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".Ussd.State")
	if err != nil {
		return 0, err
	}
	return Modem3gppUssdSessionState(variant.Value().(uint32)), nil
}

func (m *Modem) USSDNetworkRequest() (string, error) {
	variant, err := m.dbusObject.GetProperty(Modem3GPPInterface + ".Ussd.NetworkRequest")
	return variant.Value().(string), err
}
