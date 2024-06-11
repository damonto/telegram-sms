package modem

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/maltegrosse/go-modemmanager"
)

var (
	ErrNoATPortFound = errors.New("no at port found")

	defaultRebootCommand = "AT+CFUN=1,1"
	rebootCommands       = map[string]string{
		"quectel": "AT+QPOWD=1",
		"fibocom": "AT+CPWROFF",
		"simcom":  "AT+CPOF",
	}
)

func (m *Modem) Lock() {
	m.lock.Lock()
}

func (m *Modem) Unlock() {
	m.lock.Unlock()
}

func (m *Modem) Restart() error {
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(usbDevice, os.O_RDWR, 0666)
	if err != nil {
		slog.Error("failed to open file", "error", err)
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(m.rebootCommand() + "\r\n"); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	b := make([]byte, 16)
	if _, err := f.Read(b); err != nil && err != io.EOF {
		return err
	}
	if bytes.Contains(b, []byte("ERR")) {
		return errors.New(string(b))
	}
	return nil
}

func (m *Modem) rebootCommand() string {
	model, err := m.GetModel()
	if err != nil {
		return defaultRebootCommand
	}

	for k, v := range rebootCommands {
		if strings.Contains(strings.ToLower(model), k) {
			return v
		}
	}
	return defaultRebootCommand
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

func (m *Modem) GetICCID() (string, error) {
	sim, err := m.modem.GetSim()
	if err != nil {
		return "", err
	}
	return sim.GetSimIdentifier()
}

func (m *Modem) GetSignalQuality() (uint32, error) {
	percent, _, err := m.modem.GetSignalQuality()
	if err != nil {
		return 0, err
	}
	return percent, err
}
