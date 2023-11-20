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
	"runtime"
	"strings"
	"sync"

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
	ListNotifications() ([]Notification, error)
	ProcessNotification(seqNumber string) error
}

type lpacPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type es9pError struct {
	SubjectCode       string `json:"subjectCode"`
	ReasonCode        string `json:"reasonCode"`
	SubjectIdentifier string `json:"subjectIdentifier"`
	Message           string `json:"message"`
}

type errorResponse struct {
	Payload struct {
		lpacPayload
		Data string
	} `json:"payload"`
}

type infoResponse struct {
	Payload struct {
		lpacPayload
		Data struct {
			Eid         string `json:"eid"`
			DefaultSmds string `json:"default_smds"`
			DefaultSmdp string `json:"default_smdp"`
		} `json:"data"`
	} `json:"payload"`
}

type profileResponse struct {
	Payload struct {
		lpacPayload
		Data []Profile `json:"data"`
	} `json:"payload"`
}

type Profile struct {
	Iccid           string `json:"iccid"`
	ProviderName    string `json:"serviceProviderName"`
	ProfileName     string `json:"profileName"`
	ProfileNickname string `json:"profileNickname"`
	State           int    `json:"profileState"`
}

type notificationResponse struct {
	Payload struct {
		lpacPayload
		Data []Notification `json:"data"`
	} `json:"payload"`
}

type Notification struct {
	SeqNumber                  int    `json:"seqNumber"`
	ProfileManagementOperation int    `json:"profileManagementOperation"`
	NotificationAddress        string `json:"notificationAddress"`
	Iccid                      string `json:"iccid"`
}

type esim struct {
	device   string
	lpacPath string
	mux      sync.Mutex
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
	return fs.WalkDir(embededLpac, "lpac/linux-"+runtime.GOARCH, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, err := embededLpac.ReadFile(path)
			if err != nil {
				return err
			}

			targetPath := filepath.Join(e.lpacPath, d.Name())
			err = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return err
			}
			return os.WriteFile(targetPath, content, 0755)
		}
		return nil
	})
}

func (e *esim) clean() error {
	return os.RemoveAll(e.lpacPath)
}

func (e *esim) execute(arguments []string) ([]byte, error) {
	if !e.mux.TryLock() {
		return nil, errors.New("already in use")
	}
	defer e.mux.Unlock()

	lpacBin := e.lpacPath + "/lpac"

	os.Setenv("AT_DEVICE", e.device)
	os.Setenv("APDU_INTERFACE", e.lpacPath+"/libapduinterface_at.so")
	os.Setenv("HTTP_INTERFACE", e.lpacPath+"/libhttpinterface_curl.so")

	slog.Info("command executing", "arguments", strings.Join(arguments, " "))
	cmd := exec.Command(lpacBin, arguments...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	slog.Info("command executed", "output", stdout.String(), "stderr", stderr.String(), "error", err)

	if err != nil {
		var errResp errorResponse
		json.Unmarshal(stdout.Bytes(), &errResp)
		if errResp.Payload.Message != "" {
			var es9pErr es9pError
			json.Unmarshal([]byte(errResp.Payload.Data), &es9pErr)
			if es9pErr.Message != "" {
				return nil, fmt.Errorf("%s / %s / %s", es9pErr.SubjectCode, es9pErr.SubjectIdentifier, es9pErr.Message)
			}
		}
		return nil, errors.New(errResp.Payload.Message)
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

	resp := &infoResponse{}
	if err := json.Unmarshal(output, resp); err != nil {
		return "", err
	}
	if resp.Payload.Message != "success" {
		return "", fmt.Errorf("failed to get eid %v", err)
	}
	return resp.Payload.Data.Eid, nil
}

func (e *esim) ListProfiles() ([]Profile, error) {
	e.release()
	defer e.clean()
	output, err := e.execute([]string{"profile", "list"})
	if err != nil {
		return nil, err
	}

	resp := &profileResponse{}
	if err := json.Unmarshal(output, resp); err != nil {
		return nil, err
	}
	return resp.Payload.Data, nil
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

	os.Setenv("SMDP", smdp)
	os.Setenv("MATCHINGID", activationCode)
	os.Setenv("IMEI", imei)

	if confirmationCode != "" {
		os.Setenv("CONFIRMATION_CODE", confirmationCode)
	}

	slog.Info("downloading new eSIM profile", "smdp", smdp, "activationCode", activationCode, "confirmationCode", confirmationCode, "imei", imei)
	_, err := e.execute([]string{"download"})
	return err
}

func (e *esim) ListNotifications() ([]Notification, error) {
	e.release()
	defer e.clean()
	output, err := e.execute([]string{"notification", "list"})
	if err != nil {
		return nil, err
	}

	resp := &notificationResponse{}
	if err := json.Unmarshal(output, resp); err != nil {
		return nil, err
	}
	return resp.Payload.Data, nil
}

func (e *esim) ProcessNotification(seqNumber string) error {
	e.release()
	defer e.clean()
	_, err := e.execute([]string{"notification", "process", seqNumber})
	return err
}
