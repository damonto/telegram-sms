package modem

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type AT struct{ f *os.File }

func NewAT(device string) (*AT, error) {
	var at AT
	var err error
	at.f, err = os.OpenFile(device, os.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return nil, err
	}
	if err := at.setTermios(); err != nil {
		return nil, err
	}
	return &at, nil
}

func (a *AT) setTermios() error {
	fd := int(a.f.Fd())
	oldTermios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return err
	}
	defer unix.IoctlSetTermios(fd, unix.TCSETS, oldTermios)
	t := unix.Termios{
		Ispeed: unix.B9600,
		Ospeed: unix.B9600,
	}
	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	t.Oflag &^= unix.OPOST
	t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	t.Cflag &^= unix.CSIZE | unix.PARENB
	t.Cflag |= unix.CS8
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0
	return unix.IoctlSetTermios(fd, unix.TCSETS, &t)
}

func (a *AT) Run(command string) (string, error) {
	if _, err := a.f.WriteString(command + "\r\n"); err != nil {
		return "", err
	}
	reader := bufio.NewReader(a.f)
	var sb strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.Contains(line, "OK"):
			return strings.TrimSpace(sb.String()), nil
		case strings.Contains(line, "ERR"):
			return "", errors.New(line)
		default:
			sb.WriteString(line + "\n")
		}
	}
}

func (a *AT) Support(command string) bool {
	_, err := a.Run(command)
	return err == nil
}

func (a *AT) Close() error {
	return a.f.Close()
}

type ATCommand interface {
	Run(command []byte) ([]byte, error)
}

// region CSIM

type CSIM struct{ at *AT }

func NewCSIM(at *AT) ATCommand { return &CSIM{at: at} }

func (c *CSIM) Run(command []byte) ([]byte, error) {
	cmd := fmt.Sprintf("%X", command)
	cmd = fmt.Sprintf("AT+CSIM=%d,\"%s\"", len(cmd), cmd)
	slog.Debug("[AT] CSIM Sending", "command", cmd)
	response, err := c.at.Run(cmd)
	slog.Debug("[AT] CSIM Received", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	sw, err := c.sw(response)
	if err != nil {
		return nil, err
	}
	if sw[len(sw)-2] != 0x61 && sw[len(sw)-2] != 0x90 {
		return sw, fmt.Errorf("unexpected response: %X", sw)
	}
	if sw[len(sw)-2] == 0x61 {
		return c.read(sw[1:])
	}
	return sw, nil
}

func (c *CSIM) read(length []byte) ([]byte, error) {
	return c.Run(append([]byte{0x00, 0xC0, 0x00, 0x00}, length...))
}

func (c *CSIM) sw(sw string) ([]byte, error) {
	lastIdx := strings.LastIndex(sw, ",")
	if lastIdx == -1 {
		return nil, errors.New("invalid response")
	}
	return hex.DecodeString(sw[lastIdx+2 : len(sw)-1])
}

// endregion

// region CRSM

type CRSM struct{ at *AT }

func NewCRSM(at *AT) ATCommand { return &CRSM{at: at} }

type CRSMInstruction uint16

const (
	CRSMReadBinary   CRSMInstruction = 0xB0
	CRSMReadRecord   CRSMInstruction = 0xB2
	CRSMGetResponse  CRSMInstruction = 0xC0
	CRSMUpdateBinary CRSMInstruction = 0xD6
	CRSMUpdateRecord CRSMInstruction = 0xDC
	CRSMStatus       CRSMInstruction = 0xF2
)

type CRSMCommand struct {
	Instruction CRSMInstruction
	FileID      uint16
	P1          byte
	P2          byte
	Data        []byte
}

func (c CRSMCommand) Bytes() []byte {
	return fmt.Appendf(nil, "%d,%d,%d,%d,%d,\"%X\"", c.Instruction, c.FileID, c.P1, c.P2, len(c.Data), c.Data)
}

func (c *CRSM) Run(command []byte) ([]byte, error) {
	cmd := fmt.Sprintf("AT+CRSM=%s", command)
	slog.Debug("[AT] CRSM Sending", "command", cmd)
	response, err := c.at.Run(cmd)
	slog.Debug("[AT] CRSM Received", "response", response, "error", err)
	if err != nil {
		return nil, err
	}
	return c.sw(response)
}

func (r *CRSM) sw(sw string) ([]byte, error) {
	if !strings.Contains(sw, "+CRSM: 144") {
		return nil, fmt.Errorf("unexpected response: %s", sw)
	}
	data := strings.Replace(sw, "+CRSM: 144,0,", "", 1)
	return hex.DecodeString(data[1 : len(data)-1])
}

// endregion
