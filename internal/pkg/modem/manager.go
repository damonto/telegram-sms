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
	mutex   sync.Mutex
	IsEuicc bool
	modem   modemmanager.Modem
}

type Manager struct {
	mmgr   modemmanager.ModemManager
	modems map[string]*Modem
	reboot chan struct{}
}

var instance *Manager

func NewManager() (*Manager, error) {
	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		return nil, err
	}
	instance = &Manager{
		mmgr:   mmgr,
		modems: make(map[string]*Modem),
		reboot: make(chan struct{}, 1),
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
	modems, err := m.mmgr.GetModems()
	if err != nil {
		return err
	}

	modemAdded := false
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
		if exist, ok := m.modems[modemId]; ok {
			if exist.modem.GetObjectPath() == mm.GetObjectPath() {
				continue
			}
		}
		slog.Info("new modem added", "modemId", modemId, "objectPath", mm.GetObjectPath())
		modemAdded = true
		nm := &Modem{
			modem: mm,
		}
		nm.IsEuicc = nm.isEuicc()
		m.modems[modemId] = nm
	}
	// If the modem is not in the list, add it, and send a signal to reboot the subscriber
	if modemAdded {
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
