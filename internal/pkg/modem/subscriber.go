package modem

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
)

type MessagingSubscriber = func(modem *Modem, sms modemmanager.Sms)

func (m *Manager) SubscribeMessaging(subscriber MessagingSubscriber) {
Subscriber:
	stopChans := make([]chan struct{}, len(m.modems))
	for _, modem := range m.modems {
		stopChan := make(chan struct{}, 1)
		stopChans = append(stopChans, stopChan)
		go m.messagingSubscriber(modem, stopChan, subscriber)
	}

	<-m.rebootSignal
	slog.Info("got reboot signal, restarting messaging subscriber")
	for _, stopChan := range stopChans {
		stopChan <- struct{}{}
	}
	goto Subscriber
}

func (m *Manager) messagingSubscriber(modem *Modem, stopChan chan struct{}, subscriber MessagingSubscriber) error {
	messaging, err := modem.modem.GetMessaging()
	if err != nil {
		return err
	}

	slog.Info("subscribing to sms signals", "modem", modem.modem.GetObjectPath())
	dbusConn, err := modem.systemBusPrivate()
	if err != nil {
		return err
	}
	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", modemmanager.ModemMessagingSignalAdded, fmt.Sprint(modem.modem.GetObjectPath())),
	)
	sigChan := make(chan *dbus.Signal, 10)
	dbusConn.Signal(sigChan)

	for {
		select {
		case smsSignal := <-sigChan:
			sms, _, err := messaging.ParseAdded(smsSignal)
			if err != nil {
				slog.Error("failed to parse sms signal", "error", err)
				continue
			}

			for {
				time.Sleep(100 * time.Millisecond)
				state, err := sms.GetState()
				if err != nil {
					slog.Error("failed to get sms state", "error", err)
					break
				}
				if state == modemmanager.MmSmsStateSending || state == modemmanager.MmSmsStateSent {
					break
				}
				if state == modemmanager.MmSmsStateReceived {
					subscriber(modem, sms)
					break
				}
			}
		case <-stopChan:
			messaging.Unsubscribe()
			return nil
		}
	}
}
