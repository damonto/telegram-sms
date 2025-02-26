package lpa

import (
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
	"github.com/damonto/euicc-go/driver"
	sgp22http "github.com/damonto/euicc-go/http"
	"github.com/damonto/euicc-go/lpa"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type LPA struct {
	*lpa.Client
	modem       *modem.Modem
	transmitter driver.Transmitter
	mutex       sync.Mutex
}

type Info struct {
	EID          string
	FreeSpace    int32
	Manufacturer string
	Certificates []string
	Product      *Product
}

type Product struct {
	Country      string
	Manufacturer string
	Brand        string
}

func NewLPA(m *modem.Modem) (*LPA, error) {
	if m.PortType() != modem.ModemPortTypeQmi {
		return nil, errors.ErrUnsupported
	}
	port, _ := m.Port(modem.ModemPortTypeQmi)
	channel, err := driver.NewQMI(port, int(m.PrimarySimSlot))
	if err != nil {
		return nil, err
	}
	transmitter, err := driver.NewTransmitter(channel, 240)
	if err != nil {
		return nil, err
	}
	client := &lpa.Client{
		HTTP: &sgp22http.Client{
			Client:        driver.NewHTTPClient(30 * time.Second),
			AdminProtocol: "gsma/rsp/v2.2.0",
		},
		APDU: transmitter,
	}
	return &LPA{modem: m, transmitter: transmitter, Client: client}, nil
}

func (l *LPA) Info() (*Info, error) {
	defer l.transmitter.Close()
	var info Info

	eid, err := l.EID()
	if err != nil {
		return nil, err
	}
	info.EID = hex.EncodeToString(eid)
	country, manufacturer, productName := util.LookupEUM(info.EID)
	info.Product = &Product{
		Country:      country,
		Manufacturer: manufacturer,
		Brand:        productName,
	}

	tlv, err := l.EUICCInfo2()
	if err != nil {
		return nil, err
	}

	// sasAcreditationNumber
	info.Manufacturer = string(tlv.First(bertlv.Universal.Primitive(12)).Value)

	// euiccCiPKIdListForSigning
	for _, child := range tlv.First(bertlv.ContextSpecific.Constructed(10)).Children {
		info.Certificates = append(info.Certificates, util.FindCertificateIssuer(hex.EncodeToString(child.Value)))
	}

	// extResource.freeNonVolatileMemory
	resource := tlv.First(bertlv.ContextSpecific.Primitive(4))
	if resource == nil {
		return nil, errors.New("resource not found")
	}
	resource.ParseChildren()
	primitive.UnmarshalInt(&info.FreeSpace).UnmarshalBinary(resource.First(bertlv.ContextSpecific.Primitive(2)).Value)
	return &info, nil
}

func (l *LPA) Download(ac string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return nil
}
