package modem

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/maltegrosse/go-modemmanager"
	"github.com/pkg/term"
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

func (m *Modem) isEuicc() bool {
	m.Lock()
	defer m.Unlock()

	m.RunATCommand("AT+CCHC=1")
	m.RunATCommand("AT+CCHC=2")
	m.RunATCommand("AT+CCHC=3")

	response, err := m.RunATCommand("AT+CCHO=\"A0000005591010FFFFFFFF8900000100\"")
	if err != nil || !strings.Contains(response, "+CCHO") {
		slog.Error("failed to open ISD-R channel", "error", err, "response", response)
		return false
	}
	channelId := regexp.MustCompile(`\d+`).FindString(response)
	if _, err = m.RunATCommand(fmt.Sprintf("AT+CCHC=%s", channelId)); err != nil {
		slog.Error("failed to close ISD-R channel", "error", err)
		return false
	}
	return true
}

func (m *Modem) Restart() error {
	m.Lock()
	defer m.Unlock()
	response, err := m.RunATCommand(m.rebootCommand())
	if !strings.Contains(response, "OK") {
		return errors.New("failed to reboot modem" + response)
	}
	return err
}

func (m *Modem) RunATCommand(command string) (string, error) {
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return "", err
	}

	t, err := term.Open(usbDevice, term.Speed(115200), term.RawMode)
	if err != nil {
		return "", err
	}
	t.SetReadTimeout(200 * time.Millisecond)
	defer t.Close()

	var data bytes.Buffer
	buf := bufio.NewWriter(&data)
	slog.Debug("running AT command", "command", command)
	if _, err := t.Write([]byte(command + "\r\n")); err != nil {
		return "", err
	}
	if _, err = io.Copy(buf, t); err != nil {
		return "", err
	}
	buf.Flush()
	t.Flush()
	t.Close()

	response := strings.Replace(data.String(), command, "", 1)
	slog.Debug("AT command response", "command", command, "response", response)
	return response, nil
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
