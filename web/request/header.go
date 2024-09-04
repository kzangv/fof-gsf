package request

import (
	"net/http"
	"net/textproto"
	"reflect"
)

type _HeaderSource map[string][]string

func (hs _HeaderSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt _SetOptions) (isSetted bool, err error) {
	return _ValueSet(value, field, hs, textproto.CanonicalMIMEHeaderKey(tagValue), opt)
}

type HeaderBind struct{}

func (HeaderBind) Bind(req *http.Request, obj interface{}) error {
	if _, err := _Mapping(reflect.ValueOf(obj), _EmptyField, _HeaderSource(req.Header), "header"); err != nil {
		return err
	}
	return Validate(obj)
}
