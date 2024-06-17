package util

//go:generate curl -o eum.json https://euicc-manual.osmocom.org/docs/pki/eum/manifest.json
//go:generate curl -o ci.json https://euicc-manual.osmocom.org/docs/pki/ci/manifest.json

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
)

//go:embed eum.json
var eum []byte

//go:embed ci.json
var ci []byte

type EUM struct {
	EUM          string    `json:"eum"`
	Country      string    `json:"country"`
	Manufacturer string    `json:"manufacturer"`
	Products     []Product `json:"products"`
}

type Product struct {
	Pattern string `json:"pattern"`
	Name    string `json:"name"`
}

type CertificateIssuer struct {
	KeyID   string `json:"key-id"`
	Country string `json:"country"`
	Name    string `json:"name"`
}

var certificateIssuers []*CertificateIssuer
var EUMs []*EUM

func init() {
	if err := json.Unmarshal(eum, &EUMs); err != nil {
		slog.Error("failed to unmarshal EUMs", err)
	}
	if err := json.Unmarshal(ci, &certificateIssuers); err != nil {
		slog.Error("failed to unmarshal certificate issuers", err)
	}
}

func FincCertificateIssuer(keyID string) string {
	for _, ci := range certificateIssuers {
		if strings.HasPrefix(keyID, ci.KeyID) {
			return ci.Name
		}
	}
	return keyID
}

func MatchEUM(eid string) (string, string, string) {
	var country, manufacturer, productName string
	for _, manifest := range EUMs {
		if strings.HasPrefix(eid, manifest.EUM) {
			country = manifest.Country
			manufacturer = manifest.Manufacturer
			for _, product := range manifest.Products {
				if match, err := filepath.Match(product.Pattern, eid); err != nil || !match {
					continue
				}
				productName = product.Name
			}
		}
	}
	return country, manufacturer, productName
}
