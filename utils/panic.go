package utils

import "fmt"

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func Panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
