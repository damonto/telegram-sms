package util

//go:generate curl -o eum.json https://euicc-manual.osmocom.org/docs/pki/eum/manifest-v2.json
//go:generate curl -o ci.json https://euicc-manual.osmocom.org/docs/pki/ci/manifest.json

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
)

//go:embed eum.json
var eum []byte

//go:embed ci.json
var ci []byte

//go:embed sas-up-accredited-sites.json
var sasUpAccreditedSites []byte

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

type SasUpAccreditedSite struct {
	Supplier      string `json:"supplier"`
	Site          string `json:"site"`
	FirstYear     int    `json:"first_year"`
	ValidTo       int    `json:"valid_to"`
	CertificateNo string `json:"certificate_no"`
}

var (
	certificateIssuers   []*CertificateIssuer
	EUMs                 []*EUM
	SasUpAccreditedSites []*SasUpAccreditedSite
)

func init() {
	if err := json.Unmarshal(eum, &EUMs); err != nil {
		slog.Error("Failed to unmarshal EUMs", "error", err)
	}
	if err := json.Unmarshal(ci, &certificateIssuers); err != nil {
		slog.Error("Failed to unmarshal certificate issuers", "error", err)
	}
	if err := json.Unmarshal(sasUpAccreditedSites, &SasUpAccreditedSites); err != nil {
		slog.Error("Failed to unmarshal SAS UP accredited sites", "error", err)
	}
}

func FindSasUpAccreditedSite(certificateNo string) *SasUpAccreditedSite {
	prefix := certificateNo[:8]
	for _, site := range SasUpAccreditedSites {
		if strings.HasPrefix(site.CertificateNo, prefix) {
			return site
		}
	}
	return nil
}

func FindCertificateIssuer(keyID string) string {
	for _, ci := range certificateIssuers {
		if strings.HasPrefix(keyID, ci.KeyID) {
			return ci.Name
		}
	}
	return keyID
}

func LookupEUM(eid string) (country string, manufacturer string, brand string) {
	for _, manifest := range EUMs {
		if strings.HasPrefix(eid, manifest.EUM) {
			country = manifest.Country
			manufacturer = manifest.Manufacturer
			for _, product := range manifest.Products {
				if strings.HasPrefix(eid, product.Prefix) {
					if product.InRange != nil {
						eidRange, _ := strconv.Atoi(eid[len(product.Prefix):30])
						for _, inRange := range product.InRange {
							if eidRange >= inRange[0] && eidRange <= inRange[1] {
								return country, manufacturer, product.Name
							}
						}
					}
					brand = product.Name
				}
			}
		}
	}
	return country, manufacturer, brand
}
