package server

import (
	"fmt"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
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

func NewError(code int, message string) error {
	return errors.WithStack(&Error{Code: code, Message: message})
}

func ErrorBadRequest(format string, a ...interface{}) error {
	return errors.WithStack(&govalidator.Error{
		Name: "参数错误",
		Err:  fmt.Errorf(format, a...),
	})
}

func ErrorNoArgument() error {
	return ErrorBadRequest("请传入参数")
}
