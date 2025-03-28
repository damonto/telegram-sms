package lpa

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/damonto/euicc-go/apdu"
	"github.com/damonto/euicc-go/bertlv"
	"github.com/damonto/euicc-go/bertlv/primitive"
	"github.com/damonto/euicc-go/driver/at"
	"github.com/damonto/euicc-go/driver/mbim"
	"github.com/damonto/euicc-go/driver/qmi"
	"github.com/damonto/euicc-go/lpa"
	sgp22 "github.com/damonto/euicc-go/v2"
	"github.com/damonto/telegram-sms/internal/pkg/config"
	"github.com/damonto/telegram-sms/internal/pkg/modem"
	"github.com/damonto/telegram-sms/internal/pkg/util"
)

type LPA struct {
	*lpa.Client
	mutex sync.Mutex
}

type Info struct {
	EID                   string
	FreeSpace             int32
	SasAcreditationNumber string
	Manufacturer          string
	Certificates          []string
	Product               *Product
}

type Product struct {
	Country      string
	Manufacturer string
	Brand        string
}

var AIDs = [][]byte{
	lpa.GSMAISDRApplicationAID,
	{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x05, 0x05, 0x00}, // 5ber Ultra
	{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0x00, 0x00, 0x00, 0x89, 0x00, 0x00, 0x00, 0x03, 0x00}, // eSIM.me V2
	{0xA0, 0x65, 0x73, 0x74, 0x6B, 0x6D, 0x65, 0xFF, 0xFF, 0xFF, 0xFF, 0x49, 0x53, 0x44, 0x2D, 0x52}, // ESTKme 2025
	{0xA0, 0x00, 0x00, 0x05, 0x59, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x00, 0x01, 0x77}, // XeSIM
	{0xA0, 0x00, 0x00, 0x06, 0x28, 0x10, 0x10, 0xFF, 0xFF, 0xFF, 0xFF, 0x89, 0x00, 0x00, 0x01, 0x00}, // GlocalMe
}

func New(m *modem.Modem) (*LPA, error) {
	var l = new(LPA)
	ch, err := l.createChannel(m)
	if err != nil {
		return nil, err
	}
	opt := &lpa.Option{
		Channel: ch,
		MSS:     util.If(config.C.Slowdown, 120, 250),
	}
	if err := l.tryCreateClient(opt); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *LPA) tryCreateClient(opt *lpa.Option) error {
	var err error
	for _, opt.AID = range AIDs {
		l.Client, err = lpa.New(opt)
		if err == nil {
			slog.Info("LPA client created", "AID", fmt.Sprintf("%X", opt.AID))
			return nil
		}
		slog.Warn("Failed to create LPA client", "AID", fmt.Sprintf("%X", opt.AID), "error", err)
	}
	return errors.New("no supported ISD-R AID found or it's not an eUICC")
}

func (l *LPA) createChannel(m *modem.Modem) (apdu.SmartCardChannel, error) {
	slot := uint8(util.If(m.PrimarySimSlot > 0, m.PrimarySimSlot, 1))
	var err error
	switch m.PrimaryPortType() {
	case modem.ModemPortTypeQmi:
		slog.Info("Using QMI driver", "port", m.PrimaryPort, "slot", slot)
		return qmi.New(m.PrimaryPort, slot, true)
	case modem.ModemPortTypeMbim:
		slog.Info("Using MBIM driver", "port", m.PrimaryPort, "slot", slot)
		return mbim.New(m.PrimaryPort, slot, true)
	default:
		var port *modem.ModemPort
		if port, err = m.Port(modem.ModemPortTypeAt); err != nil {
			return nil, err
		}
		slog.Info("Using AT driver", "port", port.Device, "slot", slot)
		return at.New(port.Device)
	}
}

func (l *LPA) Close() error {
	return l.Client.Close()
}

func (l *LPA) Info() (*Info, error) {
	var info Info
	eid, err := l.EID()
	if err != nil {
		return nil, err
	}
	info.EID = hex.EncodeToString(eid)
	country, manufacturer, brand := util.LookupEUM(info.EID)
	info.Product = &Product{Country: country, Manufacturer: manufacturer, Brand: brand}

	tlv, err := l.EUICCInfo2()
	if err != nil {
		return nil, err
	}
	// sasAcreditationNumber
	info.SasAcreditationNumber = string(tlv.First(bertlv.Universal.Primitive(12)).Value)
	if site := util.FindSasUpAccreditedSite(info.SasAcreditationNumber); site != nil {
		info.Manufacturer = site.Supplier
	}
	// euiccCiPKIdListForSigning
	for _, child := range tlv.First(bertlv.ContextSpecific.Constructed(10)).Children {
		info.Certificates = append(info.Certificates, util.FindCertificateIssuer(hex.EncodeToString(child.Value)))
	}
	// extResource.freeNonVolatileMemory
	resource := tlv.First(bertlv.ContextSpecific.Primitive(4))
	data, _ := resource.MarshalBinary()
	data[0] = 0x30
	if err := resource.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	primitive.UnmarshalInt(&info.FreeSpace).UnmarshalBinary(resource.First(bertlv.ContextSpecific.Primitive(2)).Value)
	return &info, nil
}

func (l *LPA) Delete(id sgp22.ICCID) (sgp22.SequenceNumber, error) {
	if err := l.DeleteProfile(id); err != nil {
		return 0, err
	}
	return l.sendNotification(id, sgp22.NotificationEventDelete)
}

func (l *LPA) SendNotification(seq sgp22.SequenceNumber) error {
	ns, err := l.RetrieveNotificationList(seq)
	if err != nil {
		return err
	}
	if len(ns) > 0 {
		return l.HandleNotification(ns[0])
	}
	return nil
}

func (l *LPA) sendNotification(id sgp22.ICCID, event sgp22.NotificationEvent) (sgp22.SequenceNumber, error) {
	ln, err := l.ListNotification(event)
	if err != nil {
		return 0, err
	}
	var latest sgp22.SequenceNumber
	for _, n := range ln {
		if n.SequenceNumber > latest && n.ICCID.String() == id.String() {
			latest = n.SequenceNumber
		}
	}
	// Some profiles may not have notifications
	if latest > 0 {
		slog.Info("Sending notification", "event", event, "sequence", latest)
		ns, err := l.RetrieveNotificationList(latest)
		if err != nil {
			return 0, err
		}
		if len(ns) > 0 {
			return latest, l.HandleNotification(ns[0])
		}
	}
	return 0, nil
}

func (l *LPA) Download(ctx context.Context, activationCode *lpa.ActivationCode, handler lpa.DownloadHandler) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	slog.Info("Downloading profile", "activationCode", activationCode)
	n, err := l.DownloadProfile(ctx, activationCode, handler)
	if err != nil {
		return err
	}
	// Some profiles may not have notifications
	if n.Notification.SequenceNumber > 0 {
		slog.Info("Sending download notification", "sequence", n.Notification.SequenceNumber)
		ns, err := l.RetrieveNotificationList(n.Notification.SequenceNumber)
		if err != nil {
			return err
		}
		if len(ns) > 0 {
			return l.HandleNotification(ns[0])
		}
	}
	return nil
}
