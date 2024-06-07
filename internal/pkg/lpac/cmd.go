package lpac

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/damonto/telegram-sms/config"
)

type Cmd struct {
	ctx       context.Context
	usbDevice string
}

func NewCmd(ctx context.Context, usbDevice string) *Cmd {
	return &Cmd{ctx: ctx, usbDevice: usbDevice}
}

func (c *Cmd) Run(arguments []string, dst any, progress Progress) error {
	cmd := exec.CommandContext(c.ctx, filepath.Join(config.C.Dir, "lpac"), arguments...)
	cmd.Env = append(cmd.Env, "LPAC_APDU=at")
	cmd.Env = append(cmd.Env, "AT_DEVICE="+c.usbDevice)

	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	cmdErr := c.process(stdout, dst, progress)
	if err := cmd.Wait(); err != nil {
		slog.Error("command wait error", "error", err, "stderr", stderr.String())
	}
	return cmdErr
}

func (c *Cmd) process(output io.ReadCloser, dst any, progress Progress) error {
	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanLines)
	var cmdErr error
	for scanner.Scan() {
		if err := c.handleOutput(scanner.Text(), dst, progress); err != nil {
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
	if step, ok := HumanReadableSteps[progressPayload.Message]; ok {
		return progress(step)
	}
	return progress(progressPayload.Message)
}
