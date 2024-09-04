package request

import (
	"net/http"
	"reflect"
)

type _FormSource map[string][]string

func (form _FormSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt _SetOptions) (isSetted bool, err error) {
	return _ValueSet(value, field, form, tagValue, opt)
}

type FormBind struct {
}

func (FormBind) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}
	if err := _FormMap(formTagName, obj, req.PostForm); err != nil {
		return err
	}
	return Validate(obj)
}

func _FormParse(ptr interface{}, form map[string][]string) error {
	el := reflect.TypeOf(ptr).Elem()

	if el.Kind() == reflect.Slice {
		ptrMap, ok := ptr.(map[string][]string)
		if !ok {
			return ErrMapSlicesToStringsType
		}
		for k, v := range form {
			ptrMap[k] = v
		}

		return nil
	}

	ptrMap, ok := ptr.(map[string]string)
	if !ok {
		return ErrMapToStringsType
	}
	for k, v := range form {
		ptrMap[k] = v[len(v)-1] // pick last
	}

	return nil
}

func _FormMap(tag string, ptr interface{}, form map[string][]string) error {
	ptrVal := reflect.ValueOf(ptr)
	var pointed interface{}
	if ptrVal.Kind() == reflect.Ptr {
		ptrVal = ptrVal.Elem()
		pointed = ptrVal.Interface()
	}
	if ptrVal.Kind() == reflect.Map &&
		ptrVal.Type().Key().Kind() == reflect.String {
		if pointed != nil {
			ptr = pointed
		}
		return _FormParse(ptr, form)
	}

	_, err := _Mapping(reflect.ValueOf(ptr), _EmptyField, _FormSource(form), tag)
	return err
}
