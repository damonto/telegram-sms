package modem

import "github.com/maltegrosse/go-modemmanager"

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
	state, err := ussd.GetState()
	if err != nil {
		return "", err
	}
	if state == modemmanager.MmModem3gppUssdSessionStateActive || state == modemmanager.MmModem3gppUssdSessionStateIdle {
		if err := ussd.Cancel(); err != nil {
			return "", err
		}
	}
	reply, err := ussd.Respond(response)
	if err != nil {
		return "", err
	}
	return reply, nil
}

func (m *Modem) CancelUSSDSession() error {
	three3gpp, err := m.modem.Get3gpp()
	if err != nil {
		return err
	}
	ussd, err := three3gpp.GetUssd()
	if err != nil {
		return err
	}
	state, err := ussd.GetState()
	if err != nil {
		return err
	}
	if state == modemmanager.MmModem3gppUssdSessionStateActive || state == modemmanager.MmModem3gppUssdSessionStateIdle {
		return ussd.Cancel()
	}
	return nil
}
