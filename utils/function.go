package utils

import (
	"reflect"
	"runtime"
)

func FunctionName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
