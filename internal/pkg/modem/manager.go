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

type modem struct {
	lock  sync.Mutex
	modem modemmanager.Modem
}

type Manager struct {
	modem  *modem
	mmgr   modemmanager.ModemManager
	modems map[string]*modem
	reboot chan struct{}
}

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
		modems: make(map[string]*modem, 1),
		reboot: make(chan struct{}, 1),
	}
	go m.watch()
	return m, nil
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
		slog.Info("new modem added", "modemId", modemId, "objectPath", m.modem.modem.GetObjectPath())
		hasNewModemAdded = true
		m.modems[modemId] = &modem{
			modem: mm,
		}
	}

	// If the modem is not in the list, add it, and send a signal to reboot the subscriber
	if hasNewModemAdded {
		m.reboot <- struct{}{}
	}
	return nil
}

func (m *Manager) Modems() map[string]*modem {
	return m.modems
}

func (m *Manager) Get(modemId string) (*modem, error) {
	modem, ok := m.modems[modemId]
	if !ok {
		return nil, ErrModemNotFound
	}
	return modem, nil
}
