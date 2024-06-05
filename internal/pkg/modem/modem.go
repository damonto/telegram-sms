package modem

import (
	"errors"
	"fmt"

	"github.com/maltegrosse/go-modemmanager"
)

var (
	ErrNoATPortFound = errors.New("no at port found")
)

func (m *modem) Lock() {
	m.lock.Lock()
}

func (m *modem) Unlock() {
	m.lock.Unlock()
}

func (m *modem) GetAtPort() (string, error) {
	ports, err := m.modem.GetPorts()
	if err != nil {
		return "", err
	}

	for _, port := range ports {
		if port.PortType == modemmanager.MmModemPortTypeAt {
			return fmt.Sprintf("/dev/%s", port.PortName), nil
		}
	}
	return "", ErrNoATPortFound
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
