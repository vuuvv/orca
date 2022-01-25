package utils

import "strings"

func LineJoin(first string, second string) string {
	if strings.HasSuffix(first, "\n") {
		return first + second
	}
	return first + "\n" + second
}
