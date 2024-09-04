package request

import (
	"errors"
	"unsafe"
)

const (
	defaultMemory = 32 << 20
	formTagName   = "form"
)

type Unmarshaler interface {
	UnmarshalForm(v string) error
}

type SliceUnmarshaler interface {
	UnmarshalForm(v []string) error
}

var (
	Validate = func(obj interface{}) error { return nil }

	JsonEnableDecoderUseNumber             = false
	JsonEnableDecoderDisallowUnknownFields = false
)

var (
	ErrUnknownType            = errors.New("unknown type")
	ErrMapSlicesToStringsType = errors.New("cannot convert to map slices of strings")
	ErrMapToStringsType       = errors.New("cannot convert to map of strings")
)

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
