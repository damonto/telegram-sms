package modem

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/lpac"
	"github.com/maltegrosse/go-modemmanager"
	"golang.org/x/sys/unix"
)

var (
	ErrNoATPortFound          = errors.New("no at port found")
	ErrNoQMIDeviceFound       = errors.New("no QMI device found")
	ErrRestartCommandNotFound = errors.New("restart command not found")
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
	if config.C.APDUDriver == config.APDUDriverAT {
		usbDevice, err = m.GetAtPort()
	} else {
		usbDevice, err = m.GetQMIDevice()
	}
	if err != nil {
		slog.Error("failed to get modem port", "error", err)
		return false, ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	info, err := lpac.NewCmd(ctx, usbDevice).Info()
	if err != nil {
		slog.Error("failed to get modem info", "error", err)
		return false, ""
	}
	slog.Info("eUICC chip detected", "EID", info.EID)
	return info.EID != "", info.EID
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
	return nil
}

func (m *Modem) RunATCommand(command string) (string, error) {
	m.Lock()
	defer m.Unlock()
	usbDevice, err := m.GetAtPort()
	if err != nil {
		return "", err
	}
	port, err := os.OpenFile(usbDevice, os.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return "", err
	}
	defer port.Close()

	oldTermios, err := unix.IoctlGetTermios(int(port.Fd()), unix.TCGETS)
	if err != nil {
		return "", err
	}
	defer unix.IoctlSetTermios(int(port.Fd()), unix.TCSETS, oldTermios)

	if err := syscall.SetNonblock(int(port.Fd()), false); err != nil {
		return "", err
	}

	t := unix.Termios{
		Ispeed: unix.B19200,
		Ospeed: unix.B19200,
	}
	t.Cflag &^= unix.CSIZE | unix.PARENB | unix.CSTOPB | unix.CRTSCTS
	t.Oflag &^= unix.OPOST
	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON | unix.IGNPAR
	t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	t.Cflag |= unix.CS8 | unix.CLOCAL | unix.CREAD
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
