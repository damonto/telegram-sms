package modem

func (m *Modem) RunUSSDCommand(command string) (string, error) {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	return ussd.Initiate(command)
}

func (m *Modem) RespondUSSDCommand(response string) (string, error) {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	return ussd.Respond(response)
}
