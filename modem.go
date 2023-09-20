package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/exp/slog"
)

type SMSSubscriber = func(modem modemmanager.Modem, sms modemmanager.Sms)

type Modem interface {
	Use(modemId string) *modem
	Reload() error
	ListModems() (map[string]string, error)
	GetAtDevice() (string, error)
	GetIccid() (string, error)
	GetOperatorName() (string, error)
	GetImei() (string, error)
	GetSignalQuality() (uint32, error)
	GetPrimarySimSlot() (uint32, error)
	SetPrimarySimSlot(simSlot uint32) error
	GetSimSlots() ([]dbus.ObjectPath, error)
	RunUSSDCommand(command string) (string, error)
	RespondUSSDCommand(response string) (string, error)
	SubscribeSMS(callback SMSSubscriber)
	SendSMS(number, message string) error
}

type modem struct {
	mmgr           modemmanager.ModemManager
	modem          modemmanager.Modem
	modems         map[string]modemmanager.Modem
	modemAddedChan chan struct{}
}

var (
	modemPropertySimSlots       = "org.freedesktop.ModemManager1.Modem.SimSlots"
	modemPropertyPrimarySimSlot = "org.freedesktop.ModemManager1.Modem.PrimarySimSlot"
	modemSetPrimarySimSlot      = "org.freedesktop.ModemManager1.Modem.SetPrimarySimSlot"
)

func NewModem() (Modem, error) {
	var err error
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}
	mmgr.SetLogging(modemmanager.MMLoggingLevelWarning)
	m := &modem{
		mmgr:           mmgr,
		modemAddedChan: make(chan struct{}, 1),
	}
	if err := m.initModemManager(); err != nil {
		return nil, err
	}
	return m, err
}

func (m *modem) initModemManager() error {
	modems, err := m.mmgr.GetModems()
	if err != nil {
		return err
	}

	if len(modems) == 0 {
		return errors.New("no modems found")
	}

	mmModems := make(map[string]modemmanager.Modem, len(modems))
	for _, modem := range modems {
		state, err := modem.GetState()
		if err != nil {
			return err
		}
		if state == modemmanager.MmModemStateDisabled {
			if err := modem.Enable(); err != nil {
				slog.Error("failed to enable modem", "error", err)
				return err
			}
		}

		modemId, err := modem.GetEquipmentIdentifier()
		if err != nil {
			return err
		}
		mmModems[modemId] = modem
	}
	m.modems = mmModems
	return nil
}

func (m *modem) Use(modemId string) *modem {
	m.modem = m.modems[modemId]
	return m
}

func (m *modem) ListModems() (map[string]string, error) {
	modemList := make(map[string]string)
	for modemId, modem := range m.modems {
		threeGpp, err := modem.Get3gpp()
		if err != nil {
			return nil, err
		}
		operatorName, _ := threeGpp.GetOperatorName()
		manufacturer, _ := modem.GetManufacturer()
		hwRevision, _ := modem.GetHardwareRevision()

		modemList[modemId] = fmt.Sprintf("%s %s (Current: %s)", manufacturer, hwRevision, operatorName)
	}

	return modemList, nil
}

func (m *modem) GetAtDevice() (string, error) {
	ports, err := m.modem.GetPorts()
	if err != nil {
		return "", err
	}

	for _, port := range ports {
		if port.PortType == modemmanager.MmModemPortTypeAt {
			return fmt.Sprintf("/dev/%s", port.PortName), nil
		}
	}

	return "", errors.New("no at port founded")
}

func (m *modem) GetIccid() (string, error) {
	sim, err := m.modem.GetSim()
	if err != nil {
		return "", err
	}
	return sim.GetSimIdentifier()
}

func (m *modem) GetImei() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	return threeGpp.GetImei()
}

func (m *modem) GetOperatorName() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}

	return threeGpp.GetOperatorName()
}

func (m *modem) GetSignalQuality() (uint32, error) {
	percent, _, err := m.modem.GetSignalQuality()
	if err != nil {
		return 0, err
	}
	return percent, err
}

func (m *modem) GetSimSlots() ([]dbus.ObjectPath, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), modemPropertySimSlots)
	if err != nil {
		return nil, err
	}

	simSlots := []dbus.ObjectPath{}
	simSlots = append(simSlots, prop.Value().([]dbus.ObjectPath)...)
	return simSlots, err
}

func (m *modem) GetPrimarySimSlot() (uint32, error) {
	prop, err := m.getProperty(m.modem.GetObjectPath(), modemPropertyPrimarySimSlot)
	if err != nil {
		return 0, err
	}
	return prop.Value().(uint32), err
}

func (m *modem) SetPrimarySimSlot(simSlot uint32) error {
	primarySlot, err := m.GetPrimarySimSlot()
	if err != nil {
		return err
	}

	if primarySlot == simSlot {
		slog.Info("the given sim slot has already activated", "slot", simSlot)
		return nil
	}

	if err := m.callMethod(m.modem.GetObjectPath(), modemSetPrimarySimSlot, uint32(simSlot)); err != nil {
		return err
	}

	return m.Reload()
}

func (m *modem) Reload() error {
	for {
		time.Sleep(5 * time.Second)
		modems, err := m.mmgr.GetModems()
		if err != nil {
			return err
		}

		if len(modems) >= len(m.modems) {
			registeredModems := make(map[dbus.ObjectPath]struct{}, len(m.modems))
			for _, registeredModem := range m.modems {
				registeredModems[registeredModem.GetObjectPath()] = struct{}{}
			}

			for _, modem := range modems {
				if _, ok := registeredModems[modem.GetObjectPath()]; !ok {
					slog.Info("new modem found", "modem", modem.GetObjectPath())
					state, err := modem.GetState()
					if err != nil {
						return err
					}
					if state == modemmanager.MmModemStateDisabled {
						if err := modem.Enable(); err != nil {
							slog.Error("failed to enable modem", "error", err)
							return err
						}
					}

					modemId, err := modem.GetEquipmentIdentifier()
					if err != nil {
						return err
					}

					m.modems[modemId] = modem
					slog.Info("new modem added", "modem-id", modemId, "modem", modem.GetObjectPath())
					m.modemAddedChan <- struct{}{}
					return nil
				}
			}
		}
	}
}

func (m *modem) callMethod(objectPath dbus.ObjectPath, method string, args ...interface{}) error {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return err
	}

	obj := dbusConn.Object(modemmanager.ModemManagerInterface, objectPath)
	return obj.Call(method, 0, args...).Err
}

func (m *modem) getProperty(objectPath dbus.ObjectPath, property string) (dbus.Variant, error) {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return dbus.Variant{}, err
	}

	obj := dbusConn.Object(modemmanager.ModemManagerInterface, objectPath)
	return obj.GetProperty(property)
}

func (m *modem) RunUSSDCommand(command string) (string, error) {
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

func (m *modem) RespondUSSDCommand(response string) (string, error) {
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

func (m *modem) SendSMS(number, message string) error {
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

func (m *modem) SubscribeSMS(subscriber SMSSubscriber) {
Subscriber:
	stopChans := make([]chan struct{}, 0, len(m.modems))
	for modemId, modem := range m.modems {
		stopCh := make(chan struct{})
		stopChans = append(stopChans, stopCh)
		go m.subscribe(modemId, modem, subscriber, stopCh)
	}

	<-m.modemAddedChan
	slog.Info("unsubscribe all existing sms event")
	for _, stopChan := range stopChans {
		stopChan <- struct{}{}
	}
	goto Subscriber
}

func (m *modem) subscribe(modemId string, modem modemmanager.Modem, subscriber SMSSubscriber, stopCh chan struct{}) error {
	messaging, err := modem.GetMessaging()
	if err != nil {
		slog.Error("failed to get messaging", "error", err)
	}
	slog.Info("subscribe new sms event", "modem-id", modemId, "path", messaging.GetObjectPath())

	dbusConn, err := m.systemBusPrivate()
	if err != nil {
		return err
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", modemmanager.ModemMessagingSignalAdded, fmt.Sprint(modem.GetObjectPath()))
	dbusConn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)
	sigChan := make(chan *dbus.Signal, 10)
	dbusConn.Signal(sigChan)

	for {
		select {
		case smsSignal := <-sigChan:
			sms, _, err := messaging.ParseAdded(smsSignal)
			if err != nil {
				slog.Error("failed to parse SMS", "error", err)
				continue
			}

			slog.Info("new sms received", "path", sms.GetObjectPath(), "messaging-path", messaging.GetObjectPath())
			state, err := sms.GetState()
			if err != nil {
				slog.Error("failed to get SMS state", "error", err)
				continue
			}

			if state == modemmanager.MmSmsStateReceiving {
				for {
					time.Sleep(1 * time.Second)
					if state, err := sms.GetState(); state == modemmanager.MmSmsStateReceived && err == nil {
						subscriber(modem, sms)
						break
					}
				}
			}

			if state == modemmanager.MmSmsStateReceived {
				subscriber(modem, sms)
			} else {
				continue
			}
		case <-stopCh:
			slog.Info("unsubscribe messaging", "modem-id", modemId, "path", messaging.GetObjectPath())
			messaging.Unsubscribe()
			return nil
		}
	}
}

func (m *modem) systemBusPrivate() (*dbus.Conn, error) {
	dbusConn, err := dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}

	err = dbusConn.Auth(nil)
	if err != nil {
		dbusConn.Close()
		return nil, err
	}

	err = dbusConn.Hello()
	if err != nil {
		dbusConn.Close()
		return nil, err
	}

	return dbusConn, nil
}
