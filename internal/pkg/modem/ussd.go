package modem

func (m *Manager) RunUSSDCommand(command string) (string, error) {
	three3gpp, err := m.modem.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	return ussd.Initiate(command)
}

func (m *Manager) RespondUSSDCommand(response string) (string, error) {
	three3gpp, err := m.modem.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	return ussd.Respond(response)
}
