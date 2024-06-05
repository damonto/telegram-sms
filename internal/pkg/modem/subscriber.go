package modem

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
)

type SMSSubscriber = func(modem modemmanager.Modem, sms modemmanager.Sms)

func (m *Manager) SubscribeSMS(subscriber SMSSubscriber) {
Subscriber:
	stopChans := make([]chan struct{}, len(m.modems))
	for _, modem := range m.modems {
		stopChan := make(chan struct{}, 1)
		stopChans = append(stopChans, stopChan)
		go m.smsSubscriber(modem, stopChan, subscriber)
	}

	<-m.reboot
	slog.Info("rebooting sms subscriber")
	for _, stopChan := range stopChans {
		stopChan <- struct{}{}
	}
	goto Subscriber
}

func (m *Manager) smsSubscriber(modem *modem, stopChan chan struct{}, subscriber SMSSubscriber) error {
	messaging, err := modem.modem.GetMessaging()
	if err != nil {
		return err
	}

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

			state, err := sms.GetState()
			if err != nil {
				slog.Error("failed to get sms state", "error", err)
				continue
			}

			if state == modemmanager.MmSmsStateReceiving {
				for {
					time.Sleep(1 * time.Second)
					if state, err := sms.GetState(); state == modemmanager.MmSmsStateReceived && err == nil {
						break
					}
				}
			}

			if state == modemmanager.MmSmsStateReceived {
				subscriber(modem.modem, sms)
			}
			continue
		case <-stopChan:
			messaging.Unsubscribe()
			return nil
		}
	}
}
