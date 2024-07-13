package modem

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/sys/unix"
)

var (
	ErrNoATPortFound          = errors.New("no at port found")
	ErrRestartCommandNotFound = errors.New("restart command not found")

	restartCommands = map[string]string{
		"quectel": "AT+QPOWD=1",
		"fibocom": "AT+CPWROFF",
		"simcom":  "AT+CPOF",
	}
)

func (m *Modem) Lock() {
	m.mutex.Lock()
}

func (m *Modem) Unlock() {
	m.mutex.Unlock()
}

func (m *Modem) isEuicc() bool {
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
	model, err := m.GetModel()
	if err != nil {
		return err
	}
	manufacturer, err := m.GetManufacturer()
	if err != nil {
		return err
	}
	var restartCommand string
	for brand, command := range restartCommands {
		if strings.Contains(strings.ToLower(model), brand) || strings.Contains(strings.ToLower(manufacturer), brand) {
			restartCommand = command
			break
		}
	}
	if restartCommand == "" {
		return ErrRestartCommandNotFound
	}
	_, err = m.RunATCommand(restartCommand)
	return err
}

func (m *Modem) RunATCommand(command string) (string, error) {
	m.Lock()
	defer m.Unlock()

	usbDevice, err := m.GetAtPort()
	if err != nil {
		return "", err
	}
	port, err := os.OpenFile(usbDevice, os.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return "", err
	}
	defer port.Close()

	oldTermios, err := unix.IoctlGetTermios(int(port.Fd()), unix.TCGETS)
	if err != nil {
		return "", err
	}
	defer unix.IoctlSetTermios(int(port.Fd()), unix.TCSETS, oldTermios)

	t := unix.Termios{
		Ispeed: unix.B19200,
		Ospeed: unix.B19200,
	}
	t.Cflag &^= unix.CSIZE | unix.PARENB | unix.CSTOPB
	t.Cflag |= unix.CS8 | unix.CLOCAL | unix.CREAD
	t.Iflag &^= unix.IXON | unix.IXOFF | unix.IXANY | unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL
	t.Lflag &^= unix.ICANON | unix.ECHO | unix.ECHOE | unix.ECHONL | unix.ISIG
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(int(port.Fd()), unix.TCSETS, &t); err != nil {
		return "", err
	}

	slog.Debug("running AT command", "command", command)
	if _, err := port.WriteString(command + "\r\n"); err != nil {
		return "", err
	}

	reader := bufio.NewReader(port)
	var response string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.Contains(line, "OK") {
			break
		}
		if strings.Contains(line, "ERR") {
			return "", errors.New("failed to execute AT command: " + line)
		}
		response += line
	}
	response = strings.Replace(strings.TrimSpace(response), command, "", 1)
	slog.Debug("AT command executed", "command", command, "response", response)
	return response, nil
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
