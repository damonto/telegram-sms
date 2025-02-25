package util

//go:generate curl -L -o carrier.json https://mno-list.harded.org/unified.json

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed carrier.json
var carrier []byte

type Carrier struct {
	Brand       string              `json:"brand,omitempty"`
	Operator    string              `json:"operator,omitempty"`
	MccmncTuple map[string][]string `json:"mccmnc_tuple,omitempty"`
}

var dictionary map[string]string

func init() {
	dictionary = make(map[string]string)
	var c []Carrier
	if err := json.Unmarshal(carrier, &c); err != nil {
		panic(err)
	}
	for _, v := range c {
		for _, tuple := range v.MccmncTuple {
			for _, mccmnc := range tuple {
				if v.Brand != "" {
					dictionary[mccmnc] = fmt.Sprintf("%s - %s", v.Operator, v.Brand)
				} else {
					dictionary[mccmnc] = v.Operator
				}
			}
		}
	}
}

func LookupCarrier(mccmnc string) string {
	if operator, ok := dictionary[mccmnc]; ok {
		return operator
	}
	return "unknown"
}
