package modem

import "log/slog"

func (m *Modem) RunUSSDCommand(command string) (string, error) {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}
	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return "", err
	}
	reply, err := ussd.Initiate(command)
	if err != nil {
		if err := ussd.Cancel(); err != nil {
			slog.Error("failed to cancel ussd command", "error", err)
		}
		return "", err
	}
	return reply, nil
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
	reply, err := ussd.Respond(response)
	if err != nil {
		if err := ussd.Cancel(); err != nil {
			slog.Error("failed to cancel ussd command", "error", err)
		}
		return "", err
	}
	return reply, nil
}
