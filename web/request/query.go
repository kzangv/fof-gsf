package request

import "net/http"

type QueryBind struct {
}

func (QueryBind) Bind(req *http.Request, obj interface{}) error {
	values := req.URL.Query()
	if err := _FormMap(formTagName, obj, values); err != nil {
		return err
	}
	return Validate(obj)
}
