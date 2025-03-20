package modem

import (
	"bufio"
	"errors"
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
