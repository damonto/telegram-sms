package modem

func (m *Modem) SendSMS(number, message string) error {
	messaging, err := m.modem.GetMessaging()
	if err != nil {
		return err
	}

	sms, err := messaging.CreateSms(number, message)
	if err != nil {
		return err
	}
	return sms.Send()
}
