package modem

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/maltegrosse/go-modemmanager"
)

var (
	ErrModemNotFound = errors.New("modem not found")
)

type Modem struct {
	mutex   sync.Mutex
	IsEuicc bool
	Eid     string
	modem   modemmanager.Modem
}

type Manager struct {
	mmgr            modemmanager.ModemManager
	modems          map[string]*Modem
	resubscribeChan chan struct{}
}

var instance *Manager

func NewManager() (*Manager, error) {
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}
	instance = &Manager{
		mmgr:            mmgr,
		modems:          make(map[string]*Modem),
		resubscribeChan: make(chan struct{}, 10),
	}
	go instance.watch()
	return instance, nil
}

func GetManager() *Manager {
	return instance
}

func (m *Manager) watch() error {
	for {
		if err := m.watchModems(); err != nil {
			slog.Error("failed to watch modems", "error", err)
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) watchModems() error {
	if err := m.mmgr.ScanDevices(); err != nil {
		slog.Warn("failed to scan devices", "error", err)
	}
	modems, err := m.mmgr.GetModems()
	if err != nil {
		return err
	}
	shouldResubscriber := false
	currentModems := make(map[string]dbus.ObjectPath)
	for _, mm := range modems {
		state, err := mm.GetState()
		if err != nil {
			return err
		}
		if state == modemmanager.MmModemStateDisabled {
			if err := mm.Enable(); err != nil {
				return err
			}
		}
		modemId, err := mm.GetEquipmentIdentifier()
		if err != nil {
			return err
		}
		currentModems[modemId] = mm.GetObjectPath()
		if exist, ok := m.modems[modemId]; ok {
			if exist.modem.GetObjectPath() == mm.GetObjectPath() {
				continue
			}
		}
		slog.Info("new modem added", "modemId", modemId, "objectPath", mm.GetObjectPath())
		shouldResubscriber = true
		nm := &Modem{modem: mm}
		nm.IsEuicc, nm.Eid = nm.detectEuicc()
		m.modems[modemId] = nm
	}
	for modemId, modem := range m.modems {
		if _, ok := currentModems[modemId]; !ok {
			slog.Info("modem removed", "modemId", modemId, "objectPath", modem.modem.GetObjectPath())
			delete(m.modems, modemId)
			shouldResubscriber = true
		}
	}
	if shouldResubscriber {
		m.resubscribeChan <- struct{}{}
	}
	return nil
}

func (m *Manager) GetModems() map[string]*Modem {
	return m.modems
}

func (m *Manager) GetModem(modemId string) (*Modem, error) {
	modem, ok := m.modems[modemId]
	if !ok {
		return nil, ErrModemNotFound
	}
	return modem, nil
}
