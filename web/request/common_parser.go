package request

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	_EmptyField = reflect.StructField{}
)

type _SetOptions struct {
	isDefaultExists bool
	defaultValue    string
}

type _Setter interface {
	TrySet(value reflect.Value, field reflect.StructField, key string, opt _SetOptions) (isSetted bool, err error)
}

func head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}

func _ValueSet(value reflect.Value, field reflect.StructField, form map[string][]string, tagValue string, opt _SetOptions) (isSetted bool, err error) {
	vs, ok := form[tagValue]
	if !ok && !opt.isDefaultExists {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		if ok, err = _SliceUnmarshaler(value, vs); ok {
			return true, err
		}
		return true, _SetSlice(vs, value, field)
	case reflect.Array:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		if ok, err = _SliceUnmarshaler(value, vs); ok {
			return true, err
		}
		if len(vs) != value.Len() {
			return false, fmt.Errorf("%q is not valid value for %s", vs, value.Type().String())
		}
		return true, _SetArray(vs, value, field)
	default:
		var val string
		if !ok {
			val = opt.defaultValue
		}

		if len(vs) > 0 {
			val = vs[0]
		}
		return true, _SetValue(val, value, field)
	}
}

func _TryToSetValue(value reflect.Value, field reflect.StructField, setter _Setter, tag string) (bool, error) {
	var tagValue string
	var setOpt _SetOptions

	tagValue = field.Tag.Get(tag)
	tagValue, opts := head(tagValue, ",")

	if tagValue == "" { // default value is FieldName
		tagValue = field.Name
	}
	if tagValue == "" { // when field is "_EmptyField" variable
		return false, nil
	}

	var opt string
	for len(opts) > 0 {
		opt, opts = head(opts, ",")

		if k, v := head(opt, "="); k == "default" {
			setOpt.isDefaultExists = true
			setOpt.defaultValue = v
		}
	}

	return setter.TrySet(value, field, tagValue, setOpt)
}

func _Mapping(value reflect.Value, field reflect.StructField, setter _Setter, tag string) (bool, error) {
	if field.Tag.Get(tag) == "-" { // just ignoring this field
		return false, nil
	}

	var vKind = value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSetted, err := _Mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			return false, err
		}
		if isNew && isSetted {
			value.Set(vPtr)
		}
		return isSetted, nil
	}

	if vKind != reflect.Struct || !field.Anonymous {
		ok, err := _TryToSetValue(value, field, setter, tag)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSetted bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := _Mapping(value.Field(i), tValue.Field(i), setter, tag)
			if err != nil {
				return false, err
			}
			isSetted = isSetted || ok
		}
		return isSetted, nil
	}
	return false, nil
}

func _SetIntField(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		value.SetInt(intVal)
	}
	return err
}

func _SetUintField(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		value.SetUint(uintVal)
	}
	return err
}

func _SetBoolField(val string, value reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		value.SetBool(boolVal)
	}
	return err
}

func _SetFloatField(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		value.SetFloat(floatVal)
	}
	return err
}

func _SetTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixnano":
		tv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}

		d := time.Duration(1)
		if tf == "unixnano" {
			d = time.Second
		}

		t := time.Unix(tv/int64(d), tv%int64(d))
		value.Set(reflect.ValueOf(t))
		return nil

	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

func _SetTimeDuration(val string, value reflect.Value) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}

func _SetArray(vals []string, value reflect.Value, field reflect.StructField) error {
	for i, s := range vals {
		err := _SetValue(s, value.Index(i), field)
		if err != nil {
			return err
		}
	}
	return nil
}

func _SetSlice(vals []string, value reflect.Value, field reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := _SetArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

func _SliceUnmarshaler(v reflect.Value, vs []string) (bool, error) {
	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		if u, ok := v.Interface().(SliceUnmarshaler); ok {
			return true, u.UnmarshalForm(vs)
		}
		if u, ok := v.Interface().(Unmarshaler); ok {
			var val string
			if len(vs) > 0 {
				val = vs[0]
			}
			return true, u.UnmarshalForm(val)
		}
	}
	return false, nil
}

func _ValueUnmarshaler(v reflect.Value) Unmarshaler {
	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		if u, ok := v.Interface().(Unmarshaler); ok {
			return u
		}
	}
	return nil
}

func _SetValue(val string, value reflect.Value, field reflect.StructField) error {
	if u := _ValueUnmarshaler(value); u != nil {
		return u.UnmarshalForm(val)
	}
	switch value.Kind() {
	case reflect.Int:
		return _SetIntField(val, 0, value)
	case reflect.Int8:
		return _SetIntField(val, 8, value)
	case reflect.Int16:
		return _SetIntField(val, 16, value)
	case reflect.Int32:
		return _SetIntField(val, 32, value)
	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			return _SetTimeDuration(val, value)
		}
		return _SetIntField(val, 64, value)
	case reflect.Uint:
		return _SetUintField(val, 0, value)
	case reflect.Uint8:
		return _SetUintField(val, 8, value)
	case reflect.Uint16:
		return _SetUintField(val, 16, value)
	case reflect.Uint32:
		return _SetUintField(val, 32, value)
	case reflect.Uint64:
		return _SetUintField(val, 64, value)
	case reflect.Bool:
		return _SetBoolField(val, value)
	case reflect.Float32:
		return _SetFloatField(val, 32, value)
	case reflect.Float64:
		return _SetFloatField(val, 64, value)
	case reflect.String:
		value.SetString(val)
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return _SetTimeField(val, field, value)
		}
		return json.Unmarshal(StringToBytes(val), value.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal(StringToBytes(val), value.Addr().Interface())
	default:
		return ErrUnknownType
	}
	return nil
}
