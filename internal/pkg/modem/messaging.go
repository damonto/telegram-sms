package modem

import (
	"context"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
)

const ModemMessagingInterface = ModemInterface + ".Messaging"

func (m *Modem) ListMessages() ([]*SMS, error) {
	messages := new([]dbus.ObjectPath)
	err := m.dbusObject.Call(ModemMessagingInterface+".List", 0).Store(messages)
	var s []*SMS
	for _, message := range *messages {
		sms, err := m.RetrieveSMS(message)
		if err != nil {
			return nil, err
		}
		s = append(s, sms)
	}
	return s, err
}

func (m *Modem) CreateMessage(to string, text string) (dbus.ObjectPath, error) {
	var path dbus.ObjectPath
	data := map[string]any{
		"number": to,
		"text":   text,
	}
	err := m.dbusObject.Call(ModemMessagingInterface+".Create", 0, &data).Store(&path)
	return path, err
}

func (m *Modem) DeleteMessage(path dbus.ObjectPath) error {
	return m.dbusObject.Call(ModemMessagingInterface+".Delete", 0, path).Err
}

func (m *Modem) SubscribeMessaging(ctx context.Context, subscriber func(message *SMS) error) error {
	dbusConn, err := m.SystemBusPrivate()
	if err != nil {
		return err
	}
	dbusConn.AddMatchSignal(
		dbus.WithMatchMember("Added"),
		dbus.WithMatchPathNamespace(m.objectPath),
	)
	signalChan := make(chan *dbus.Signal, 10)
	dbusConn.Signal(signalChan)
	defer dbusConn.RemoveSignal(signalChan)
	for {
		select {
		case sig := <-signalChan:
			if !sig.Body[1].(bool) {
				continue
			}
			s, err := m.waitForSMSReceived(sig.Body[0].(dbus.ObjectPath))
			if err != nil {
				slog.Error("Failed to process message", "error", err, "path", sig.Path)
				continue
			}
			if err := subscriber(s); err != nil {
				slog.Error("Failed to process message", "error", err, "path", sig.Path)
			}
		case <-ctx.Done():
			slog.Info("Unsubscribing from modem messaging", "path", m.dbusObject.Path())
			return nil
		}
	}
}

func (m *Modem) waitForSMSReceived(path dbus.ObjectPath) (*SMS, error) {
	for {
		s, err := m.RetrieveSMS(path)
		if err != nil {
			return nil, err
		}
		if s.State == SMSStateReceived {
			return s, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}
