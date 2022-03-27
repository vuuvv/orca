package utils

import (
	"fmt"
	"go.uber.org/zap"
)

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func Panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

func NormalRecover(caller string) {
	if r := recover(); r != nil {
		zap.L().Error("recover from: "+caller, zap.String("error", fmt.Sprintf("%+v", r)))
	}
}

func Catch(caller string, handler func(reason interface{})) {
	if r := recover(); r != nil {
		zap.L().Error("recover from: "+caller, zap.String("error", fmt.Sprintf("%+v", r)))
		handler(r)
	}
}
