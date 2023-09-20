package esim

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog"
)

type Esim interface {
	Eid() (string, error)
	ListProfiles() ([]Profile, error)
	Download(smdp string, activationCode string, confirmationCode string, imei string) error
	Rename(iccid string, name string) error
	Enable(iccid string) error
	Disable(iccid string) error
	Delete(iccid string) error
}

type commandResponse struct {
	Message string `json:"message"`
}

type es9pError struct {
	SubjectCode       string `json:"subjectCode"`
	ReasonCode        string `json:"reasonCode"`
	SubjectIdentifier string `json:"subjectIdentifier"`
	Message           string `json:"message"`
}

type errorResponse struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

type Eid struct {
	Eid string `json:"eid"`
}

type Profile struct {
	Iccid           string `json:"iccid"`
	ProviderName    string `json:"serviceProviderName"`
	ProfileName     string `json:"profileName"`
	ProfileNickname string `json:"profileNickname"`
	State           int    `json:"profileState"`
}

type esim struct {
	device   string
	lpacPath string
}

//go:embed lpac/*
var embededLpac embed.FS

func New(device string) Esim {
	return &esim{
		device: device,
	}
}

func (e *esim) release() error {
	var err error
	e.lpacPath, err = os.MkdirTemp("", "lpac")
	if err != nil {
		return err
	}
	return fs.WalkDir(embededLpac, "lpac", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, err := embededLpac.ReadFile(path)
			if err != nil {
				return err
			}

			targetPath := filepath.Join(e.lpacPath, path)
			err = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return err
			}
			err = os.WriteFile(targetPath, content, 0644)
			if err != nil {
				return err
			}

			return os.Chmod(targetPath, 0755)
		}
		return nil
	})
}

func (e *esim) clean() error {
	return os.RemoveAll(e.lpacPath)
}

func (e *esim) execute(arguments []string) ([]byte, error) {
	lpacBin := e.lpacPath + "/lpac/lpac"

	os.Setenv("AT_DEVICE", e.device)
	os.Setenv("APDU_INTERFACE", e.lpacPath+"/lpac/libapduinterface_at.so")
	os.Setenv("ES9P_INTERFACE", e.lpacPath+"/lpac/libes9pinterface_curl.so")
	os.Setenv("OUTPUT_JSON", "1")

	slog.Info("command executing", "arguments", strings.Join(arguments, " "))
	cmd := exec.Command(lpacBin, arguments...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	slog.Info("command executed", "output", stdout.String(), "error", err)

	if err != nil {
		var errResp errorResponse
		json.Unmarshal(stdout.Bytes(), &errResp)
		if errResp.Data != "" {
			var es9pErr es9pError
			json.Unmarshal([]byte(errResp.Data), &es9pErr)
			if es9pErr.Message != "" {
				return nil, errors.New(es9pErr.Message)
			}
		}
		return nil, errors.New(errResp.Message)
	}

	return stdout.Bytes(), nil
}

func (e *esim) Eid() (string, error) {
	e.release()
	defer e.clean()
	output, err := e.execute([]string{"info"})
	if err != nil {
		return "", err
	}

	type response struct {
		commandResponse
		Data Eid `json:"data"`
	}
	resp := &response{}
	if err := json.Unmarshal(output, resp); err != nil {
		return "", err
	}
	if resp.Message != "success" {
		return "", fmt.Errorf("failed to get eid %v", err)
	}
	return resp.Data.Eid, nil
}

func (e *esim) ListProfiles() ([]Profile, error) {
	e.release()
	defer e.clean()
	output, err := e.execute([]string{"profile", "list"})
	if err != nil {
		return nil, err
	}

	type response struct {
		commandResponse
		Data []Profile `json:"data"`
	}
	resp := &response{}
	if err := json.Unmarshal(output, resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (e *esim) Rename(iccid string, name string) error {
	e.release()
	defer e.clean()
	_, err := e.execute([]string{"profile", "rename", iccid, name})
	return err
}

func (e *esim) Enable(iccid string) error {
	e.release()
	defer e.clean()
	_, err := e.execute([]string{"profile", "enable", iccid})
	return err
}

func (e *esim) Disable(iccid string) error {
	e.release()
	defer e.clean()
	_, err := e.execute([]string{"profile", "disable", iccid})
	return err
}

func (e *esim) Delete(iccid string) error {
	e.release()
	defer e.clean()
	_, err := e.execute([]string{"profile", "delete", iccid})
	return err
}

func (e *esim) Download(smdp string, activationCode string, confirmationCode string, imei string) error {
	e.release()
	defer e.clean()

	args := []string{"download", "-a", smdp, "-m", activationCode, "-i", imei}
	if confirmationCode != "" {
		args = append(args, "-c", confirmationCode)
	}
	_, err := e.execute(args)
	return err
}
