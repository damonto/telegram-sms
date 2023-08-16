package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"time"
)

// Paths of methods and properties
const (
	SmsInterface = ModemManagerInterface + ".Sms"

	/* Methods */
	SmsSend  = SmsInterface + ".Send"
	SmsStore = SmsInterface + ".Store"

	/* Property */
	SmsPropertyState   = SmsInterface + ".State"   // readable   u
	SmsPropertyPduType = SmsInterface + ".PduType" //  readable   u
	SmsPropertyNumber  = SmsInterface + ".Number"  // readable   s
	SmsPropertyText    = SmsInterface + ".Text"    // readable   s
	SmsPropertyData    = SmsInterface + ".Data"    // readable   ay

	SmsPropertySMSC            = SmsInterface + ".SMSC"            // readable   s
	SmsPropertyValidity        = SmsInterface + ".Validity"        // readable   (uv)
	SmsPropertyClass           = SmsInterface + ".Class"           // readable   i
	SmsPropertyTeleserviceId   = SmsInterface + ".TeleserviceId"   // readable   u
	SmsPropertyServiceCategory = SmsInterface + ".ServiceCategory" // readable   u

	SmsPropertyDeliveryReportRequest = SmsInterface + ".DeliveryReportRequest" // readable   b
	SmsPropertyMessageReference      = SmsInterface + ".MessageReference"      // readable   u
	SmsPropertyTimestamp             = SmsInterface + ".Timestamp"             // readable   s
	SmsPropertyDischargeTimestamp    = SmsInterface + ".DischargeTimestamp"    //  readable   s
	SmsPropertyDeliveryState         = SmsInterface + ".DeliveryState"         // readable   u

	SmsPropertyStorage = SmsInterface + ".Storage" // readable   u

)

// The Sms interface Defines operations and properties of a single SMS message.
type Sms interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// If the message has not yet been sent, queue it for delivery.
	Send() error

	// This method requires a MMSmsStorage value, describing the storage where this message is to be kept; or
	// MM_SMS_STORAGE_UNKNOWN if the default storage should be used.
	Store(MMSmsStorage) error

	/* PROPERTIES */

	// A MMSmsState value, describing the state of the message.
	GetState() (MMSmsState, error)

	// A MMSmsPduType value, describing the type of PDUs used in the SMS message.
	GetPduType() (MMSmsPduType, error)

	// Number to which the message is addressed.
	GetNumber() (string, error)

	// Message text, in UTF-8.
	// When sending, if the text is larger than the limit of the technology or modem, the message will be broken
	// into multiple parts or messages.
	// Note that Text and Data are never given at the same time.
	GetText() (string, error)

	// Message data.
	// When sending, if the data is larger than the limit of the technology or modem, the message will be broken
	// into multiple parts or messages.
	// Note that Text and Data are never given at the same time.
	GetData() ([]byte, error)

	// Indicates the SMS service center number.
	// Always empty for 3GPP2/CDMA.
	GetSMSC() (string, error)

	// Indicates when the SMS expires in the SMSC.
	// This value is composed of a MMSmsValidityType key, with an associated data which contains type-specific validity information:
	// 		MM_SMS_VALIDITY_TYPE_RELATIVE: The value is the length of the validity period in minutes, given as an unsigned integer (D-Bus signature 'u').
	GetValidity() (map[MMSmsValidityType]interface{}, error)

	// 3GPP message class (-1..3). -1 means class is not available or is not used for this message, otherwise the 3GPP SMS message class.
	// Always -1 for 3GPP2/CDMA.
	GetClass() (int32, error)

	// A MMSmsCdmaTeleserviceId value.
	// Always MM_SMS_CDMA_TELESERVICE_ID_UNKNOWN for 3GPP.
	GetTeleserviceId() (MMSmsCdmaTeleserviceId, error)

	// A MMSmsCdmaServiceCategory value.
	// Always MM_SMS_CDMA_SERVICE_CATEGORY_UNKNOWN for 3GPP.
	GetServiceCategory() (MMSmsCdmaServiceCategory, error)

	// #TRUE if delivery report request is required, #FALSE otherwise.
	GetDeliveryReportRequest() (bool, error)

	// Message Reference of the last PDU sent/received within this SMS.
	// If the PDU type is MM_SMS_PDU_TYPE_STATUS_REPORT, this field identifies the Message Reference of the
	// PDU associated to the status report.
	GetMessageReference() (MMSmsPduType, error)

	// Time when the first PDU of the SMS message arrived the SMSC, in ISO8601 format. This field is only applicable if
	// the PDU type is MM_SMS_PDU_TYPE_DELIVER. or MM_SMS_PDU_TYPE_STATUS_REPORT.
	GetTimestamp() (time.Time, error)

	// Time when the first PDU of the SMS message left the SMSC, in ISO8601 format.
	// This field is only applicable if the PDU type is MM_SMS_PDU_TYPE_STATUS_REPORT.
	GetDischargeTimestamp() (time.Time, error)

	// A MMSmsDeliveryState value, describing the state of the delivery reported in the Status Report message.
	// This field is only applicable if the PDU type is MM_SMS_PDU_TYPE_STATUS_REPORT.
	GetDeliveryState() (MMSmsDeliveryState, error)

	// A MMSmsStorage value, describing the storage where this message is kept.
	GetStorage() (MMSmsStorage, error)

	MarshalJSON() ([]byte, error)

	/* SIGNALS */

	// Listen to changed properties
	// returns []interface
	// index 0 = name of the interface on which the properties are defined
	// index 1 = changed properties with new values as map[string]dbus.Variant
	// index 2 = invalidated properties: changed properties but the new values are not send with them
	SubscribePropertiesChanged() <-chan *dbus.Signal

	// ParsePropertiesChanged parses the dbus signal
	ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error)
	Unsubscribe()
}

// NewSms returns new Sms Interface
func NewSms(objectPath dbus.ObjectPath) (Sms, error) {
	var ss sms
	return &ss, ss.init(ModemManagerInterface, objectPath)
}

type sms struct {
	dbusBase
	sigChan chan *dbus.Signal
}

func (ss sms) GetObjectPath() dbus.ObjectPath {
	return ss.obj.Path()
}

func (ss sms) Send() error {
	return ss.call(SmsSend)
}

func (ss sms) Store(MMSmsStorage) error {
	return ss.call(SmsStore)
}

func (ss sms) GetState() (MMSmsState, error) {
	res, err := ss.getUint32Property(SmsPropertyState)
	if err != nil {
		return MmSmsStateUnknown, err
	}
	return MMSmsState(res), nil
}

func (ss sms) GetPduType() (MMSmsPduType, error) {
	res, err := ss.getUint32Property(SmsPropertyPduType)
	if err != nil {
		return MmSmsPduTypeUnknown, err
	}
	return MMSmsPduType(res), nil
}

func (ss sms) GetNumber() (string, error) {
	return ss.getStringProperty(SmsPropertyNumber)

}

func (ss sms) GetText() (string, error) {
	return ss.getStringProperty(SmsPropertyText)
}

func (ss sms) GetData() ([]byte, error) {
	// todo untested
	return ss.getSliceByteProperty(SmsPropertyData)
}

func (ss sms) GetSMSC() (string, error) {
	return ss.getStringProperty(SmsPropertySMSC)
}

func (ss sms) GetValidity() (map[MMSmsValidityType]interface{}, error) {
	// todo: untested
	res, err := ss.getInterfaceProperty(SmsPropertyValidity)
	if err != nil {
		return nil, err
	}
	var myMap map[MMSmsValidityType]interface{}
	myMap = make(map[MMSmsValidityType]interface{})
	result, ok := res.([]interface{})
	if ok {
		for key, element := range result {
			myMap[MMSmsValidityType(key)] = element
		}
	}
	return myMap, nil

}

func (ss sms) GetClass() (int32, error) {
	return ss.getInt32Property(SmsPropertyClass)
}

func (ss sms) GetTeleserviceId() (MMSmsCdmaTeleserviceId, error) {
	res, err := ss.getUint32Property(SmsPropertyTeleserviceId)
	if err != nil {
		return MmSmsCdmaTeleserviceIdUnknown, err
	}

	return MMSmsCdmaTeleserviceId(res), nil
}

func (ss sms) GetServiceCategory() (MMSmsCdmaServiceCategory, error) {
	res, err := ss.getUint32Property(SmsPropertyServiceCategory)
	if err != nil {
		return MmSmsCdmaServiceCategoryUnknown, err
	}
	return MMSmsCdmaServiceCategory(res), nil
}

func (ss sms) GetDeliveryReportRequest() (bool, error) {
	return ss.getBoolProperty(SmsPropertyDeliveryReportRequest)
}

func (ss sms) GetMessageReference() (MMSmsPduType, error) {
	res, err := ss.getUint32Property(SmsPropertyMessageReference)
	if err != nil {
		return MmSmsPduTypeUnknown, err
	}
	return MMSmsPduType(res), nil
}

func (ss sms) GetTimestamp() (time.Time, error) {
	res, err := ss.getStringProperty(SmsPropertyTimestamp)
	if err != nil {
		return time.Now(), err
	}
	if len(res) < 1 {
		return time.Now(), errors.New("no timestamp available")
	}
	t, err := time.Parse(time.RFC3339Nano, res)
	if err != nil {
		return time.Now(), err
	}
	return t, err
}

func (ss sms) GetDischargeTimestamp() (time.Time, error) {
	res, err := ss.getStringProperty(SmsPropertyDischargeTimestamp)
	if err != nil {
		return time.Now(), err
	}

	if len(res) < 1 {
		return time.Now(), errors.New("no discharge timestamp available")
	}
	t, err := time.Parse(time.RFC3339Nano, res)
	if err != nil {
		return time.Now(), err
	}
	return t, err
}

func (ss sms) GetDeliveryState() (MMSmsDeliveryState, error) {
	res, err := ss.getUint32Property(SmsPropertyDeliveryState)
	if err != nil {
		return MmSmsDeliveryStateUnknown, err
	}
	return MMSmsDeliveryState(res), nil
}

func (ss sms) GetStorage() (MMSmsStorage, error) {
	res, err := ss.getUint32Property(SmsPropertyStorage)
	if err != nil {
		return MmSmsStorageUnknown, err
	}
	return MMSmsStorage(res), nil
}

func (ss sms) SubscribePropertiesChanged() <-chan *dbus.Signal {
	if ss.sigChan != nil {
		return ss.sigChan
	}
	rule := fmt.Sprintf("type='signal', member='%s',path_namespace='%s'", dbusPropertiesChanged, fmt.Sprint(ss.GetObjectPath()))
	ss.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
	ss.sigChan = make(chan *dbus.Signal, 10)
	ss.conn.Signal(ss.sigChan)
	return ss.sigChan
}
func (ss sms) ParsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	return ss.parsePropertiesChanged(v)
}

func (ss sms) Unsubscribe() {
	ss.conn.RemoveSignal(ss.sigChan)
	ss.sigChan = nil
}

func (ss sms) MarshalJSON() ([]byte, error) {
	state, err := ss.GetState()
	if err != nil {
		return nil, err
	}
	pduType, err := ss.GetPduType()
	if err != nil {
		return nil, err
	}
	number, err := ss.GetNumber()
	if err != nil {
		return nil, err
	}
	text, err := ss.GetText()
	if err != nil {
		return nil, err
	}
	data, err := ss.GetData()
	if err != nil {
		return nil, err
	}
	smsc, err := ss.GetSMSC()
	if err != nil {
		return nil, err
	}
	validity, err := ss.GetValidity()
	if err != nil {
		return nil, err
	}
	class, err := ss.GetClass()
	if err != nil {
		return nil, err
	}
	teleserviceId, err := ss.GetTeleserviceId()
	if err != nil {
		return nil, err
	}
	serviceCategory, err := ss.GetServiceCategory()
	if err != nil {
		return nil, err
	}
	deliveryReportRequest, err := ss.GetDeliveryReportRequest()
	if err != nil {
		return nil, err
	}
	messageReference, err := ss.GetMessageReference()
	if err != nil {
		return nil, err
	}
	// either discharge or timestamp is set
	var timestamp time.Time
	var dischargeTimestamp time.Time
	timestamp, err = ss.GetTimestamp()
	if err != nil {
		dischargeTimestamp, _ = ss.GetDischargeTimestamp()
	}
	dischargeTimestamp, err = ss.GetDischargeTimestamp()
	if err != nil {
		timestamp, _ = ss.GetTimestamp()
	}

	deliveryState, err := ss.GetDeliveryState()
	if err != nil {
		return nil, err
	}
	storage, err := ss.GetStorage()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"State":                 state,
		"PduType":               pduType,
		"Number":                number,
		"Text":                  text,
		"Data":                  data,
		"Smsc":                  smsc,
		"Validity":              validity,
		"Class":                 class,
		"TeleserviceId":         teleserviceId,
		"ServiceCategory":       serviceCategory,
		"DeliveryReportRequest": deliveryReportRequest,
		"MessageReference":      messageReference,
		"Timestamp":             timestamp,
		"DischargeTimestamp":    dischargeTimestamp,
		"DeliveryState":         deliveryState,
		"Storage":               storage,
	})
}
