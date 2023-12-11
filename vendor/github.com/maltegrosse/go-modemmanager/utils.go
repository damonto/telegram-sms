package modemmanager

/*
 Copied and extended from https://github.com/Wifx/gonetworkmanager/blob/master/utils.go
 on April the 1st 2020
 The MIT License (MIT)
 Copyright (c) 2019 Wifx Sàrl & Copyright (c) 2016 Bellerophon Mobile
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:
 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.
 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"net"
	"reflect"
	"strings"
	"time"
)

const (
	dbusMethodAddMatch       = "org.freedesktop.DBus.AddMatch"
	dbusMethodManagedObjects = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"
	dbusPropertiesChanged    = "PropertiesChanged"
)

// Pair represents two interface values (left and right side)
type Pair struct {
	a, b interface{}
}

// GetLeft returns left value
func (p Pair) GetLeft() interface{} {
	return p.a
}

// GetRight returns right value
func (p Pair) GetRight() interface{} {
	return p.b
}

// SetLeft value
func (p Pair) SetLeft(left interface{}) {
	p.a = left
}

// SetRight value
func (p Pair) SetRight(right interface{}) {
	p.b = right
}
func (p Pair) String() string {
	return fmt.Sprint(p.a) + " : " + fmt.Sprint(p.b)
}
func (p Pair) pairToSlice() []interface{} {
	var sl []interface{}
	sl = append(sl, p.a)
	sl = append(sl, p.b)
	return sl
}

// MarshalJSON returns a byte array
func (p Pair) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Left":   p.GetLeft(),
		"Right":  p.GetRight(),
	})
}

type dbusBase struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

func (d *dbusBase) init(iface string, objectPath dbus.ObjectPath) error {
	var err error

	d.conn, err = dbus.SystemBus()
	if err != nil {
		return err
	}

	d.obj = d.conn.Object(iface, objectPath)

	return nil
}

func (d *dbusBase) call(method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Err
}

func (d *dbusBase) callWithReturn(ret interface{}, method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Store(ret)
}

func (d *dbusBase) callWithReturn2(ret1 interface{}, ret2 interface{}, method string, args ...interface{}) error {
	return d.obj.Call(method, 0, args...).Store(ret1, ret2)
}

func (d *dbusBase) subscribe(iface, member string) {
	rule := fmt.Sprintf("type='signal',interface='%s',path='%s',member='%s'",
		iface, d.obj.Path(), ModemManagerInterface)
	d.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
}

func (d *dbusBase) subscribeNamespace(namespace string) {
	rule := fmt.Sprintf("type='signal',path_namespace='%s'", namespace)
	d.conn.BusObject().Call(dbusMethodAddMatch, 0, rule)
}
func (d *dbusBase) parsePropertiesChanged(v *dbus.Signal) (interfaceName string, changedProperties map[string]dbus.Variant, invalidatedProperties []string, err error) {
	if len(v.Body) != 3 {
		err = errors.New("error by parsing property changed signal")
		return
	}
	interfaceName, ok := v.Body[0].(string)
	if !ok {
		err = errors.New("error by parsing interface name")
		return
	}
	changedProperties, ok = v.Body[1].(map[string]dbus.Variant)
	if !ok {
		err = errors.New("error by parsing changed properties map name")
		return
	}
	invalidatedProperties, ok = v.Body[2].([]string)
	if !ok {
		err = errors.New("error by parsing invalidated properties")
		return
	}
	return
}

func (d *dbusBase) getProperty(iface string) (interface{}, error) {
	variant, err := d.obj.GetProperty(iface)
	return variant.Value(), err
}

func (d *dbusBase) setProperty(iface string, value interface{}) error {
	err := d.obj.SetProperty(iface, dbus.MakeVariant(value))
	return err
}

func (d *dbusBase) getObjectProperty(iface string) (value dbus.ObjectPath, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(dbus.ObjectPath)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceObjectProperty(iface string) (value []dbus.ObjectPath, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}

	value, ok := prop.([]dbus.ObjectPath)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getBoolProperty(iface string) (value bool, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(bool)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getStringProperty(iface string) (value string, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(string)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceStringProperty(iface string) (value []string, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]string)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceSliceByteProperty(iface string) (value [][]byte, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([][]byte)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}
func (d *dbusBase) getPairProperty(iface string) (value Pair, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	values, ok := prop.([]interface{})
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	for idy, val := range values {
		if idy == 0 {
			value.a = val
		}
		if idy == 1 {
			value.b = val
		}
	}
	return
}

func (d *dbusBase) getSliceSlicePairProperty(iface string) (value []Pair, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	values, ok := prop.([][]interface{})
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	for _, xs := range values {

		var tmpPair Pair
		for idy, val := range xs {
			if idy == 0 {
				tmpPair.a = val
			}
			if idy == 1 {
				tmpPair.b = val
			}
		}
		value = append(value, tmpPair)
	}

	return
}

func (d *dbusBase) getMapStringVariantProperty(iface string) (value map[string]dbus.Variant, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(map[string]dbus.Variant)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getMapUint32Uint32Property(iface string) (value map[uint32]uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(map[uint32]uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}
func (d *dbusBase) getMapUint32InterfaceProperty(iface string) (value map[uint32]interface{}, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}

	value, ok := prop.(map[uint32]interface{})
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}
func (d *dbusBase) getMapUint32VariantProperty(iface string) (value map[uint32]dbus.Variant, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}

	value, ok := prop.(map[uint32]dbus.Variant)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getTimestampProperty(iface string) (value time.Time, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	parsedValue, ok := prop.([]interface{})
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	var sec uint64
	sec = 0
	var msec uint64
	msec = 0
	for idx, val := range parsedValue {
		if idx == 0 {
			sec, ok = val.(uint64)
			if !ok {
				err = makeErrVariantType(iface)
				return
			}
		}
		if idx == 1 {
			msec, ok = val.(uint64)
			if !ok {
				err = makeErrVariantType(iface)
				return
			}
		}
	}
	value = time.Unix(int64(sec), int64(msec))
	return
}

func (d *dbusBase) getUint8Property(iface string) (value uint8, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint8)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getUint32Property(iface string) (value uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getInterfaceProperty(iface string) (value interface{}, err error) {
	return d.getProperty(iface)
}

func (d *dbusBase) getInt32Property(iface string) (value int32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(int32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getInt64Property(iface string) (value int64, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(int64)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}
func (d *dbusBase) getFloat32Property(iface string) (value float32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(float32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getFloat64Property(iface string) (value float64, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(float64)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getUint64Property(iface string) (value uint64, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.(uint64)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceUint32Property(iface string) (value []uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}

	value, ok := prop.([]uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceSliceUint32Property(iface string) (value [][]uint32, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([][]uint32)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceMapStringVariantProperty(iface string) (value []map[string]dbus.Variant, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]map[string]dbus.Variant)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceMapStringInterfaceProperty(iface string) (value []map[string]interface{}, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]map[string]interface{})
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func (d *dbusBase) getSliceByteProperty(iface string) (value []byte, err error) {
	prop, err := d.getProperty(iface)
	if err != nil {
		return
	}
	value, ok := prop.([]byte)
	if !ok {
		err = makeErrVariantType(iface)
		return
	}
	return
}

func makeErrVariantType(iface string) error {
	return fmt.Errorf("unexpected variant type for '%s'", iface)
}

func ip4ToString(ip uint32) string {
	bs := []byte{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(bs, ip)
	return net.IP(bs).String()
}

func (d *dbusBase) getManagedObjects(iface string, path dbus.ObjectPath) ([]dbus.ObjectPath, error) {
	var managedObjectPaths []dbus.ObjectPath

	managedObjects := make(map[dbus.ObjectPath]interface{})

	busObject := d.conn.Object(iface, path)

	err := busObject.Call(dbusMethodManagedObjects, 0).Store(&managedObjects)
	if err != nil {
		return nil, err
	}

	for path := range managedObjects {
		managedObjectPaths = append(managedObjectPaths, path)
	}
	return managedObjectPaths, nil
}

func (d *dbusBase) Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (d *dbusBase) isEmpty(object interface{}, fieldName string) bool {
	v := reflect.ValueOf(object)
	st := reflect.TypeOf(object)
	for i := 0; i < v.NumField(); i++ {
		field := st.Field(i)
		if field.Name == fieldName && v.Field(i).IsZero() {
			return true
		}
	}
	return false
}

func returnString(object interface{}) string {
	var resSlice []string
	v := reflect.ValueOf(object)
	st := reflect.TypeOf(object)
	for i := 0; i < v.NumField(); i++ {
		field := st.Field(i)
		if !v.Field(i).IsZero() {
			tmpField := field.Name + ": " + fmt.Sprint(reflect.Indirect(v).FieldByName(field.Name))
			resSlice = append(resSlice, tmpField)
		}
	}
	return strings.Join(resSlice, ", ")
}
