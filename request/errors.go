package request

import (
	"fmt"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
	"net/http"
)

type Error struct {
	Code    int    // Code 错误代码
	Status  int    // Status http返回status
	Message string // Message 错误信息
	NeedLog bool   // NeedLog
	Data    interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Message)
}

func (e *Error) WithStack() error {
	return errors.WithStack(e)
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func ErrorBadRequest(format string, a ...interface{}) error {
	return errors.WithStack(&govalidator.Error{
		Name: "参数错误",
		Err:  fmt.Errorf(format, a...),
	})
}

func ErrorNoArgument(key string) error {
	msg := "参数错误：请传入参数"
	if key != "" {
		msg += "[" + key + "]"
	}
	return ErrorBadRequest(msg)
}

var ErrorForbidden = &Error{
	Code:    http.StatusForbidden,
	Status:  http.StatusForbidden,
	Message: "无权访问",
	NeedLog: false,
}

func NewErrorForbidden() error {
	return errors.WithStackAndSkip(ErrorForbidden, 1)
}

var ErrorUnauthorized = &Error{
	Code:    http.StatusUnauthorized,
	Status:  http.StatusUnauthorized,
	Message: "请先登录",
	NeedLog: false,
}

func NewErrorUnauthorized() error {
	return errors.WithStackAndSkip(ErrorUnauthorized, 1)
}
