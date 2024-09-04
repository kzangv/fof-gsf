package request

import (
	"errors"
	"mime/multipart"
	"net/http"
	"reflect"
)

type _MultipartRequest http.Request

func (r *_MultipartRequest) TrySet(value reflect.Value, field reflect.StructField, key string, opt _SetOptions) (isSetted bool, err error) {
	if files := r.MultipartForm.File[key]; len(files) != 0 {
		return _MultipartFormFileSet(value, field, files)
	}
	return _ValueSet(value, field, r.MultipartForm.Value, key, opt)
}

type FormMultipartBind struct{}

func (FormMultipartBind) Name() string {
	return "multipart/form-data"
}

func (FormMultipartBind) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	if _, err := _Mapping(reflect.ValueOf(obj), _EmptyField, (*_MultipartRequest)(req), "form"); err != nil {
		return err
	}
	return Validate(obj)
}

func _MultipartFormFileSet(value reflect.Value, field reflect.StructField, files []*multipart.FileHeader) (isSetted bool, err error) {
	switch value.Kind() {
	case reflect.Ptr:
		switch value.Interface().(type) {
		case *multipart.FileHeader:
			value.Set(reflect.ValueOf(files[0]))
			return true, nil
		}
	case reflect.Struct:
		switch value.Interface().(type) {
		case multipart.FileHeader:
			value.Set(reflect.ValueOf(*files[0]))
			return true, nil
		}
	case reflect.Slice:
		slice := reflect.MakeSlice(value.Type(), len(files), len(files))
		isSetted, err = _SetArrayOfMultipartFormFiles(slice, field, files)
		if err != nil || !isSetted {
			return isSetted, err
		}
		value.Set(slice)
		return true, nil
	case reflect.Array:
		return _SetArrayOfMultipartFormFiles(value, field, files)
	}
	return false, errors.New("unsupported field type for multipart.FileHeader")
}

func _SetArrayOfMultipartFormFiles(value reflect.Value, field reflect.StructField, files []*multipart.FileHeader) (isSetted bool, err error) {
	if value.Len() != len(files) {
		return false, errors.New("unsupported len of array for []*multipart.FileHeader")
	}
	for i := range files {
		setted, err := _MultipartFormFileSet(value.Index(i), field, files[i:i+1])
		if err != nil || !setted {
			return setted, err
		}
	}
	return true, nil
}
