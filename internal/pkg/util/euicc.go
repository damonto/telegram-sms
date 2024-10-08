package util

//go:generate curl -o eum.json https://euicc-manual.osmocom.org/docs/pki/eum/manifest-v2.json
//go:generate curl -o ci.json https://euicc-manual.osmocom.org/docs/pki/ci/manifest.json

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
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
	Prefix  string  `json:"prefix"`
	Chip    string  `json:"chip,omitempty"`
	Name    string  `json:"name"`
	InRange [][]int `json:"in-range,omitempty"`
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
		slog.Error("failed to unmarshal EUMs", "error", err)
	}
	if err := json.Unmarshal(ci, &certificateIssuers); err != nil {
		slog.Error("failed to unmarshal certificate issuers", "error", err)
	}
}

func FindCertificateIssuer(keyID string) string {
	for _, ci := range certificateIssuers {
		if strings.HasPrefix(keyID, ci.KeyID) {
			return ci.Name
		}
	}
	return keyID
}

func LookupEUM(eid string) (string, string, string) {
	var country, manufacturer, productName string
	for _, manifest := range EUMs {
		if strings.HasPrefix(eid, manifest.EUM) {
			country = manifest.Country
			manufacturer = manifest.Manufacturer
			for _, product := range manifest.Products {
				if strings.HasPrefix(eid, product.Prefix) {
					if product.InRange != nil {
						for _, inRange := range product.InRange {
							for _, r := range inRange {
								if strings.HasPrefix(eid[len(product.Prefix):], fmt.Sprintf("%d", r)) {
									return country, manufacturer, product.Name
								}
							}
						}
					}
					return country, manufacturer, product.Name
				}
			}
		}
	}
	return country, manufacturer, productName
}
