package response

import "fmt"

type Error interface {
	error
	Code() int
}

type ErrorDefault struct {
	ErrMsg  string
	ErrCode int
}

func (v *ErrorDefault) Error() string {
	return v.ErrMsg
}

func (v *ErrorDefault) Code() int {
	return v.ErrCode
}

type Response interface {
	SetErrCode(int) Response
	SetCustomError(Error) Response
	SetErrMsg(int, string, ...interface{}) Response
	SetData(interface{}) Response
	SetDataMsg(interface{}, string, ...interface{}) Response
	SetMeta(interface{}) Response
	AddError(...error) Response
}

func SetCode(v *int, code int) {
	*v = code
}

func SetData(v *interface{}, data interface{}) {
	*v = data
}

func SetMsg(v *string, msg string, msgArgs ...interface{}) {
	if len(msgArgs) > 0 {
		*v = fmt.Sprintf(msg, msgArgs...)
	} else {
		*v = msg
	}
}

func SetMeta(v *interface{}, meta interface{}) {
	if ShowMore {
		*v = meta
	}
}

func SetErrCode(c *int, m *string, code int) {
	SetCode(c, code)
	if msg, ok := CodeMsgMap[code]; ok {
		*m = msg
	} else {
		*m = "fail"
	}
}

func SetError(c *[]string, errs ...error) {
	if ShowMore {
		if *c != nil {
			nLen := 4
			if len(errs) > nLen {
				nLen = len(errs)
			}
			*c = make([]string, 0, nLen)
		}
		for _, err := range errs {
			if err != nil {
				*c = append(*c, err.Error())
			}
		}
	}
}

type _Default struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Meta   interface{} `json:"meta,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

var (
	CodeMsgMap = map[int]string{}
	ShowMore   = true
)

func (r *_Default) SetErrCode(code int) Response {
	SetErrCode(&r.Code, &r.Msg, code)
	return r
}

func (r *_Default) SetCustomError(err Error) Response {
	SetCode(&r.Code, err.Code())
	return r._SetMsg(err.Error())
}

func (r *_Default) SetErrMsg(code int, msg string, msgArgs ...interface{}) Response {
	SetCode(&r.Code, code)
	return r._SetMsg(msg, msgArgs...)
}

func (r *_Default) _SetMsg(msg string, msgArgs ...interface{}) Response {
	SetMsg(&r.Msg, msg, msgArgs...)
	return r
}

func (r *_Default) SetData(v interface{}) Response {
	SetData(&r.Data, v)
	return r
}

func (r *_Default) SetDataMsg(v interface{}, msg string, msgArgs ...interface{}) Response {
	SetData(&r.Data, v)
	SetMsg(&r.Msg, msg, msgArgs...)
	return r
}

func (r *_Default) SetMeta(v interface{}) Response {
	SetMeta(&r.Meta, v)
	return r
}

func (r *_Default) AddError(errs ...error) Response {
	SetError(&r.Errors, errs...)
	return r
}

func New() Response {
	return &_Default{}
}
