package util

//go:generate curl -L -o carrier_list.textpb https://android.googlesource.com/platform/packages/providers/TelephonyProvider/+/main/assets/latest_carrier_id/carrier_list.textpb?format=text

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"log/slog"
	"strings"
)

//go:embed carrier_list.textpb
var textpb []byte

type Carrier struct {
	CarrierName string
	MCCMNCs     []string
}

var carrierList []*Carrier

func init() {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(textpb)))
	if _, err := base64.StdEncoding.Decode(decoded, textpb); err != nil {
		slog.Error("failed to decode carrier list", "error", err)
	}
	var carrier *Carrier
	for _, b := range bytes.Split(decoded, []byte("\n")) {
		line := strings.TrimSpace(string(b))
		if strings.HasPrefix(line, "carrier_id {") {
			carrier = new(Carrier)
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
