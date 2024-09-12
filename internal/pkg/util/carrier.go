package util

//go:generate url -L -o carrier_list.textpb https://android.googlesource.com/platform/packages/providers/TelephonyProvider/+/main/assets/latest_carrier_id/carrier_list.textpb?format=text

import (
	_ "embed"
	"encoding/base64"
	"log/slog"
	"strings"
)

//go:embed carrier_list.textpb
var carrierListRaw []byte

type Carrier struct {
	CarrierName string
	MCCMNCs     []string
}

var carrierList []*Carrier

func init() {
	var decoded []byte
	decoded, err := base64.StdEncoding.DecodeString(string(carrierListRaw))
	if err != nil {
		slog.Error("failed to decode carrier list", "error", err)
	}

	lines := strings.Split(string(decoded), "\n")
	var carrier *Carrier
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "carrier_id {") {
			carrier = &Carrier{}
			carrierList = append(carrierList, carrier)
		}
		if strings.HasPrefix(line, "carrier_name:") {
			carrier.CarrierName = strings.ReplaceAll(strings.TrimSpace(strings.Split(line, ":")[1]), "\"", "")
		}
		if strings.HasPrefix(line, "mccmnc_tuple:") {
			carrier.MCCMNCs = append(carrier.MCCMNCs, strings.ReplaceAll(strings.TrimSpace(strings.Split(line, ":")[1]), "\"", ""))
		}
	}
}

func LookupCarrierName(code string) string {
	for _, carrier := range carrierList {
		for _, mccmnc := range carrier.MCCMNCs {
			if mccmnc == code {
				return carrier.CarrierName
			}
		}
	}
	return code
}
