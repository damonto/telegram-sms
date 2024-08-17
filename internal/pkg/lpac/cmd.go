package lpac

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/damonto/telegram-sms/internal/pkg/config"
)

var ErrDownloadCancelled = errors.New("download cancelled")

type Cmd struct {
	ctx       context.Context
	usbDevice string
	stdout    io.ReadCloser
	stdin     io.WriteCloser
}

func NewCmd(ctx context.Context, usbDevice string) *Cmd {
	return &Cmd{ctx: ctx, usbDevice: usbDevice}
}

func (c *Cmd) Run(arguments []string, dst any, progress Progress) error {
	cmd := exec.CommandContext(c.ctx, filepath.Join(config.C.Dir, "lpac"), arguments...)
	cmd.Env = append(cmd.Env, "LPAC_APDU="+config.C.APDUDriver)
	if config.C.APDUDriver == config.APDUDriverAT {
		cmd.Env = append(cmd.Env, "AT_DEVICE="+c.usbDevice)
	} else {
		cmd.Env = append(cmd.Env, "QMI_DEVICE="+c.usbDevice)
	}
	slog.Debug("running lpac command", "command", cmd.String(), "env", cmd.Env)

	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	c.stdout, _ = cmd.StdoutPipe()
	c.stdin, _ = cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	cmdErr := c.process(dst, progress)
	if err := cmd.Wait(); err != nil {
		slog.Error("command wait error", "error", err, "stderr", stderr.String())
	}
	if cmdErr != nil {
		return fmt.Errorf("%w %s", cmdErr, &stderr)
	}
	return nil
}

func (c *Cmd) process(dst any, progress Progress) error {
	scanner := bufio.NewScanner(c.stdout)
	scanner.Split(bufio.ScanLines)
	var cmdErr error
	for scanner.Scan() {
		text := scanner.Text()
		slog.Debug("lpac command output", "output", text)
		if err := c.handleOutput(text, dst, progress); err != nil {
			cmdErr = err
		}
	}
	return cmdErr
}

func (c *Cmd) handleOutput(output string, dst any, progress Progress) error {
	var commandOutput CommandOutput
	if err := json.Unmarshal([]byte(output), &commandOutput); err != nil {
		return err
	}
	switch commandOutput.Type {
	case CommandStdioLPA:
		return c.handleLPAResponse(commandOutput.Payload, dst)
	case CommandStdioProgress:
		if progress != nil {
			return c.handleProgress(commandOutput.Payload, progress)
		}
	}
	return nil
}

func (c *Cmd) handleLPAResponse(payload json.RawMessage, dst any) error {
	var lpaPayload LPAPyaload
	if err := json.Unmarshal(payload, &lpaPayload); err != nil {
		return err
	}

	if lpaPayload.Code != 0 {
		if lpaPayload.Message == LPADownloadCancelled {
			return ErrDownloadCancelled
		}
		var errorMessage string
		if err := json.Unmarshal(lpaPayload.Data, &errorMessage); err != nil {
			return errors.New(lpaPayload.Message)
		}
		if errorMessage == "" {
			return errors.New(lpaPayload.Message)
		}
		return errors.New(errorMessage)
	}

	if dst != nil {
		return json.Unmarshal(lpaPayload.Data, dst)
	}
	return nil
}

func (c *Cmd) handleProgress(payload json.RawMessage, progress Progress) error {
	var progressPayload ProgressPayload
	if err := json.Unmarshal(payload, &progressPayload); err != nil {
		return err
	}

	if progressPayload.Message == ProgressMetadataParse {
		profileMetadata := &Profile{}
		if err := json.Unmarshal(progressPayload.Data, profileMetadata); err != nil {
			return err
		}
		return progress(progressPayload.Message, profileMetadata, nil)
	}

	if progressPayload.Message == ProgressPreviewConfirm {
		downloadConfirmation := make(chan bool, 1)
		if err := progress(progressPayload.Message, nil, downloadConfirmation); err != nil {
			return err
		}
		if <-downloadConfirmation {
			c.stdin.Write([]byte("y\n"))
		} else {
			c.stdin.Write([]byte("n\n"))
		}
		return nil
	}

	if text, ok := HumanReadableText[progressPayload.Message]; ok {
		return progress(text, nil, nil)
	}
	return nil
}
