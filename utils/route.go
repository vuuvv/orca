package utils

import (
	"fmt"
	"strings"
)

func RouteKey(method string, path string) string {
	return fmt.Sprintf("%s::%s", strings.ToUpper(method), path)
}
