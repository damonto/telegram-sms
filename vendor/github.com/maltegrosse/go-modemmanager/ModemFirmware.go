package modemmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
)

// Paths of methods and properties
const (
	ModemFirmwareInterface = ModemInterface + ".Firmware"

	/* Methods */
	ModemFirmwareList   = ModemFirmwareInterface + ".List"
	ModemFirmwareSelect = ModemFirmwareInterface + ".Select"

	/* Property */
	ModemFirmwarePropertyUpdateSettings = ModemFirmwareInterface + ".UpdateSettings" // readable   (ua{sv})

)

// ModemFirmware provides access to perform different firmware-related operations in the modem,
// including listing the available firmware images in the module and selecting which of them to use.
// This interface does not provide direct access to perform firmware updates in the device. Instead, it
// exposes information about the expected firmware update method as well as method-specific details required for the
// upgrade to happen. The actual firmware upgrade may be performed via the Linux Vendor Firmware Service and the fwupd daemon.
// This interface will always be available as long a the modem is considered valid.
type ModemFirmware interface {
	/* METHODS */

	// Returns object path
	GetObjectPath() dbus.ObjectPath

	// List installed firmware images.
	// Firmware slots and firmware images are identified by arbitrary opaque strings.
	// 		List (OUT s      selected, OUT aa{sv} installed);

	List() ([]firmwareProperty, error)

	// Selects a different firmware image to use, and immediately resets the modem so that it begins using the new firmware image.
	// The method will fail if the identifier does not match any of the names returned by List(), or if the image could not be selected for some reason.
	// Installed images can be selected non-destructively.
	// 		IN s uniqueid: The unique ID of the firmware image to select.
	Select(string) error

	MarshalJSON() ([]byte, error)

	/* PROPERTIES */
	// Detailed settings that provide information about how the module should be updated.
	GetUpdateSettings() (updateSettingsProperty, error)
}

// NewModemFirmware returns new ModemFirmware Interface
func NewModemFirmware(objectPath dbus.ObjectPath) (ModemFirmware, error) {
	var fi modemFirmware
	return &fi, fi.init(ModemManagerInterface, objectPath)
}

type modemFirmware struct {
	dbusBase
}

// firmwareProperty represents all properties of a firmware
type firmwareProperty struct {
	ImageType         MMFirmwareImageType `json:"image-type"`           // (Required) Type of the firmware image, given as a MMFirmwareImageType value (signature "u"). Firmware images of type MM_FIRMWARE_IMAGE_TYPE_GENERIC will only expose only the mandatory properties.
	UniqueId          string              `json:"unique-id"`            // (Required) A user-readable unique ID for the firmware image, given as a string value (signature "s").
	GobiPriVersion    string              `json:"gobi-pri-version"`     // (Optional) The version of the PRI firmware image, in images of type MM_FIRMWARE_IMAGE_TYPE_GOBI, given as a string value (signature "s").
	GobiPriInfo       string              `json:"gobi-pri-info"`        // (Optional) Additional information of the PRI image, in images of type MM_FIRMWARE_IMAGE_TYPE_GOBI, given as a string value (signature "s").
	GobiBootVersion   string              `json:"gobi-boot-version"`    // (Optional) The boot version of the PRI firmware image, in images of type MM_FIRMWARE_IMAGE_TYPE_GOBI, given as a string value (signature "s").
	GobiPriUniqueId   string              `json:"gobi-pri-unique-id"`   // (Optional) The unique ID of the PRI firmware image, in images of type MM_FIRMWARE_IMAGE_TYPE_GOBI, given as a string value (signature "s").
	GobiModemUniqueId string              `json:"gobi-modem-unique-id"` // (Optional) The unique ID of the Modem firmware image, in images of type MM_FIRMWARE_IMAGE_TYPE_GOBI, given as a string value (signature "s").
	Selected          bool                `json:"selected"`             // Shows if certain firmware is selected
}

// MarshalJSON returns a byte array
func (fp firmwareProperty) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"ImageType":         fmt.Sprint(fp.ImageType),
		"UniqueId":          fp.UniqueId,
		"GobiPriVersion":    fp.GobiPriVersion,
		"GobiPriInfo":       fp.GobiPriInfo,
		"GobiBootVersion":   fp.GobiBootVersion,
		"GobiPriUniqueId":   fp.GobiPriUniqueId,
		"GobiModemUniqueId": fp.GobiModemUniqueId,
		"Selected":          fp.Selected,
	})
}
func (fp firmwareProperty) String() string {
	return "ImageType: " + fmt.Sprint(fp.ImageType) +
		", UniqueId: " + fp.UniqueId +
		", GobiPriVersion: " + fp.GobiPriVersion +
		", GobiPriInfo: " + fp.GobiPriInfo +
		", GobiBootVersion: " + fp.GobiBootVersion +
		", GobiPriUniqueId: " + fp.GobiPriUniqueId +
		", GobiModemUniqueId: " + fp.GobiModemUniqueId +
		", Selected: " + fmt.Sprint(fp.Selected)
}

// updateSettingsProperty represents all available update settings
type updateSettingsProperty struct {
	UpdateMethods []MMModemFirmwareUpdateMethod `json:"update-methods"` // The settings are given as a bitmask of MMModemFirmwareUpdateMethod values specifying the type of firmware update procedures
	DeviceIds     []string                      `json:"device-ids"`     // (Required) This property exposes the list of device IDs associated to a given device, from most specific to least specific. (signature 'as'). E.g. a list containing: "USB\VID_413C&PID_81D7&REV_0001", "USB\VID_413C&PID_81D7" and "USB\VID_413C"
	Version       string                        `json:"version"`        // (Required) This property exposes the current firmware version string of the module. If the module uses separate version numbers for firmware version and carrier configuration, this version string will be a combination of both, and so it may be different to the version string showed in the "Revision" property. (signature 's')
	FastbootAt    string                        `json:"fastboot-at"`    // only if update method fastboot: (Required) This property exposes the AT command that should be sent to the module to trigger a reset into fastboot mode (signature 's')
}

// MarshalJSON returns a byte array
func (us updateSettingsProperty) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"UpdateMethods": fmt.Sprint(us.UpdateMethods),
		"DeviceIds":     us.DeviceIds,
		"Version":       us.Version,
		"FastbootAt":    us.FastbootAt,
	})
}

func (us updateSettingsProperty) String() string {
	return "UpdateMethods: " + fmt.Sprint(us.UpdateMethods) +
		", DeviceIds: " + fmt.Sprint(us.DeviceIds) +
		", Version: " + us.Version +
		", FastbootAt: " + us.FastbootAt
}

func (fi modemFirmware) GetObjectPath() dbus.ObjectPath {
	return fi.obj.Path()
}

func (fi modemFirmware) Select(uid string) error {
	return fi.call(ModemFirmwareSelect, uid)
}

func (fi modemFirmware) List() (properties []firmwareProperty, err error) {
	var resMap []map[string]dbus.Variant
	var tmpString string
	err = fi.callWithReturn2(&tmpString, &resMap, ModemFirmwareList)
	if err != nil {
		return
	}
	for _, el := range resMap {
		var property firmwareProperty
		for key, element := range el {
			switch key {
			case "image-type":
				tmpValue, ok := element.Value().(uint32)
				if ok {
					property.ImageType = MMFirmwareImageType(tmpValue)
				}
			case "unique-id":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.UniqueId = tmpValue
					if tmpValue == tmpString {
						property.Selected = true
					}
				}
			case "gobi-pri-version":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.GobiPriVersion = tmpValue

				}
			case "gobi-pri-info":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.GobiPriInfo = tmpValue

				}
			case "gobi-boot-version":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.GobiBootVersion = tmpValue
				}
			case "gobi-pri-unique-id":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.GobiPriUniqueId = tmpValue
				}
			case "gobi-modem-unique-id":
				tmpValue, ok := element.Value().(string)
				if ok {
					property.GobiModemUniqueId = tmpValue
				}
			}
		}
		properties = append(properties, property)
	}
	return
}

func (fi modemFirmware) GetUpdateSettings() (property updateSettingsProperty, err error) {
	res, err := fi.getPairProperty(ModemFirmwarePropertyUpdateSettings)
	if err != nil {
		return
	}
	var tmp MMModemFirmwareUpdateMethod
	bitmask, ok := res.GetLeft().(uint32)
	if !ok {
		return property, errors.New("wrong type")
	}
	property.UpdateMethods = tmp.BitmaskToSlice(bitmask)
	resMap, ok := res.GetRight().(map[string]dbus.Variant)
	if !ok {
		return property, errors.New("wrong type")
	}
	for key, element := range resMap {
		switch key {
		case "device-ids":
			tmpValue, ok := element.Value().([]string)
			if ok {
				property.DeviceIds = tmpValue
			}
		case "version":
			tmpValue, ok := element.Value().(string)
			if ok {
				property.Version = tmpValue
			}
		case "fastboot-at":
			tmpValue, ok := element.Value().(string)
			if ok {
				property.FastbootAt = tmpValue
			}
		}

	}
	return
}

func (fi modemFirmware) MarshalJSON() ([]byte, error) {
	updateSettings, err := fi.GetUpdateSettings()
	if err != nil {
		return nil, err
	}
	updateSettingsJson, err := updateSettings.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]interface{}{
		"UpdateSettings": updateSettingsJson,
	})
}
