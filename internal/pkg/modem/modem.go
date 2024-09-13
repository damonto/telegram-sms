package modem

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/damonto/telegram-sms/internal/pkg/lpa"
	"github.com/maltegrosse/go-modemmanager"
)

var (
	ErrNoATPortFound    = errors.New("no at port found")
	ErrNoQMIDeviceFound = errors.New("no QMI device found")
)

func (m *Modem) Lock() {
	m.mutex.Lock()
}

func (m *Modem) Unlock() {
	m.mutex.Unlock()
}

func (m *Modem) detectEuicc() (bool, string) {
	m.Lock()
	defer m.Unlock()
	var usbDevice string
	var err error

	usbDevice, err = m.GetQMIDevice()
	if err != nil {
		slog.Error("failed to get modem port", "error", err)
		return false, ""
	}

	slot, err := m.GetPrimarySimSlot()
	if err != nil {
		slog.Error("failed to get primary sim slot", "error", err)
		return false, ""
	}

	l, err := lpa.New(usbDevice, int(slot))
	if err != nil {
		slog.Error("failed to create lpa", "error", err)
		return false, ""
	}
	defer l.Close()

	eid, err := l.GetEid()
	if err != nil {
		slog.Error("failed to get chip info", "error", err)
		return false, ""
	}
	slog.Info("eUICC chip detected", "eid", eid)
	return eid != "", eid
}

func (m *Modem) Restart() error {
	qmiDevice, err := m.GetQMIDevice()
	if err != nil {
		return err
	}
	simslot, err := m.GetPrimarySimSlot()
	if err != nil {
		return err
	}
	if result, err := exec.Command("qmicli", "-d", qmiDevice, "-p", fmt.Sprintf("--uim-sim-power-off=%d", simslot)).Output(); err != nil {
		slog.Error("failed to power off sim", "error", err, "result", string(result))
		return err
	}
	if result, err := exec.Command("qmicli", "-d", qmiDevice, "-p", fmt.Sprintf("--uim-sim-power-on=%d", simslot)).Output(); err != nil {
		slog.Error("failed to power on sim", "error", err, "result", string(result))
		return err
	}
	// Some older modems require disabling and enabling the modem to take effect.
	if err := m.modem.Disable(); err != nil {
		slog.Warn("failed to disable modem", "error", err)
	}
	return nil
}

func (m *Modem) GetAtPort() (string, error) {
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

func (m *Modem) GetQMIDevice() (string, error) {
	ports, err := m.modem.GetPorts()
	if err != nil {
		return "", err
	}
	for _, port := range ports {
		if port.PortType == modemmanager.MmModemPortTypeQmi {
			return fmt.Sprintf("/dev/%s", port.PortName), nil
		}
	}
	return "", ErrNoQMIDeviceFound
}

func (m *Modem) GetManufacturer() (string, error) {
	return m.modem.GetManufacturer()
}

func (m *Modem) GetIccid() (string, error) {
	sim, err := m.modem.GetSim()
	if err != nil {
		return "", err
	}
	return sim.GetSimIdentifier()
}

func (m *Modem) GetImei() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}
	return threeGpp.GetImei()
}

func (m *Modem) GetModel() (string, error) {
	return m.modem.GetModel()
}

func (m *Modem) GetRevision() (string, error) {
	return m.modem.GetRevision()
}

func (m *Modem) GetOperatorName() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}
	return threeGpp.GetOperatorName()
}

func (m *Modem) GetOperatorCode() (string, error) {
	threeGpp, err := m.modem.Get3gpp()
	if err != nil {
		return "", err
	}
	return threeGpp.GetOperatorCode()
}

func (m *Modem) GetSignalQuality() (uint32, error) {
	percent, _, err := m.modem.GetSignalQuality()
	if err != nil {
		return 0, err
	}
	return percent, err
}

func (m *Modem) GetOwnNumbers() ([]string, error) {
	return m.modem.GetOwnNumbers()
}
