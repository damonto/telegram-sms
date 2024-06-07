package modem

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/maltegrosse/go-modemmanager"
)

var (
	ErrModemNotFound = errors.New("modem not found")
)

type Modem struct {
	lock  sync.Mutex
	modem modemmanager.Modem
}

type Manager struct {
	mmgr   modemmanager.ModemManager
	modems map[string]*Modem
	reboot chan struct{}
}

var Instance *Manager

func NewManager() (*Manager, error) {
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}

	if err := mmgr.ScanDevices(); err != nil {
		return nil, err
	}

	m := &Manager{
		mmgr:   mmgr,
		modems: make(map[string]*Modem, 1),
		reboot: make(chan struct{}, 1),
	}
	go m.watch()
	Instance = m
	return m, nil
}

func GetManager() *Manager {
	return Instance
}

func (m *Manager) watch() error {
	for {
		if err := m.watchModems(); err != nil {
			slog.Error("failed to watch modems", "error", err)
			panic(err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (m *Manager) watchModems() error {
	modems, err := m.mmgr.GetModems()
	if err != nil {
		return err
	}

	hasNewModemAdded := false
	for _, mm := range modems {
		// state, err := mm.GetState()
		// if err != nil {
		// 	return err
		// }
		// if state == modemmanager.MmModemStateDisabled {
		// 	if err := mm.Enable(); err != nil {
		// 		return err
		// 	}
		// }
		modemId, err := mm.GetEquipmentIdentifier()
		if err != nil {
			return err
		}
		if exist, ok := m.modems[modemId]; ok {
			if exist.modem.GetObjectPath() == mm.GetObjectPath() {
				continue
			}
		}
		slog.Info("new modem added", "modemId", modemId, "objectPath", mm.GetObjectPath())
		hasNewModemAdded = true
		m.modems[modemId] = &Modem{
			modem: mm,
		}
	}

	// If the modem is not in the list, add it, and send a signal to reboot the subscriber
	if hasNewModemAdded {
		m.reboot <- struct{}{}
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
