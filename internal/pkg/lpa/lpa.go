package lpa

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
	"github.com/damonto/euicc-go/driver"
	sgp22http "github.com/damonto/euicc-go/http"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type LPA struct {
	*lpa.Client
	transmitter driver.Transmitter
	mutex       sync.Mutex
}

type Info struct {
	EID                   string
	FreeSpace             int32
	SasAcreditationNumber string
	Certificates          []string
	Product               *Product
}

type Product struct {
	Country      string
	Manufacturer string
	Brand        string
}

func New(m *modem.Modem) (*LPA, error) {
	var l = new(LPA)
	var err error
	l.transmitter, err = l.createTransmitter(m)
	if err != nil {
		return nil, err
	}
	l.Client = &lpa.Client{
		HTTP: &sgp22http.Client{
			Client:        driver.NewHTTPClient(30 * time.Second),
			AdminProtocol: "gsma/rsp/v2.2.0",
		},
		APDU: l.transmitter,
	}
	return l, nil
}

func (l *LPA) createTransmitter(m *modem.Modem) (driver.Transmitter, error) {
	port, err := m.Port(modem.ModemPortTypeQmi)
	if err != nil {
		return nil, err
	}
	channel, err := driver.NewQMI(port, int(m.PrimarySimSlot))
	if err != nil {
		return nil, err
	}
	return driver.NewTransmitter(channel, 240)
}

func (l *LPA) Close() error {
	return l.transmitter.Close()
}

func (l *LPA) Info() (*Info, error) {
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
	info.SasAcreditationNumber = string(tlv.First(bertlv.Universal.Primitive(12)).Value)

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

func (l *LPA) Delete(id sgp22.ICCID) error {
	if err := l.DeleteProfile(id); err != nil {
		return err
	}
	return l.sendNotification(id, sgp22.NotificationEventDelete)
}

func (l *LPA) sendNotification(id sgp22.ICCID, event sgp22.NotificationEvent) error {
	ln, err := l.ListNotification(event)
	if err != nil {
		return err
	}
	var latest sgp22.SequenceNumber
	for _, n := range ln {
		if n.SequenceNumber > latest && n.ICCID.String() == id.String() {
			latest = n.SequenceNumber
		}
	}
	slog.Info("Sending notification", "event", event, "sequence", latest)
	n, err := l.RetrieveNotificationList(latest)
	if err != nil {
		return err
	}
	return l.HandleNotification(n[0])
}

func (l *LPA) Download(ctx context.Context, activationCode *lpa.ActivationCode, handler lpa.DownloadHandler) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	slog.Info("Downloading profile", "activation", activationCode)
	n, err := l.DownloadProfile(ctx, activationCode, handler)
	if err != nil {
		return err
	}
	slog.Info("Sending download notification", "sequence", n.Notification.SequenceNumber)
	ns, err := l.RetrieveNotificationList(n.Notification.SequenceNumber)
	if err != nil {
		return err
	}
	return l.HandleNotification(ns[0])
}
