package esim

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type ESIM interface {
	Eid() (string, error)
	// Download(smds string, activationCode string, confirmationCode string, imei string) error
	// Rename(iccid string, name string) error
	// Enable(iccid string) error
	// Disable(iccid string) error
	// Delete(iccid string) error
}

type esim struct {
	lpacPath string
}

//go:embed lpac/*
var embededLpac embed.FS

func New(device string) (ESIM, error) {
	e := &esim{}
	if err := e.releaseLpac(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *esim) releaseLpac() error {
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
			return err
		}
		return nil
	})
}

func (e *esim) cleanLpac() error {
	return os.RemoveAll(e.lpacPath)
}

func (e *esim) execute(arguments []string) (string, error) {
	lpacBin := e.lpacPath + "/lpac/lpac"
	fmt.Println(lpacBin)
	return "", nil
}

func (e *esim) Eid() (string, error) {
	return "", nil
}
