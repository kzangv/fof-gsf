package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JsonBind struct{}

func (JsonBind) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return _JsonDecode(req.Body, obj)
}

func _JsonDecode(r io.Reader, obj interface{}) error {
	decoder := json.NewDecoder(r)
	if JsonEnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if JsonEnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return Validate(obj)
}
