package utils

import (
	"encoding/base64"
	"math/rand"
)

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return base64.StdEncoding.EncodeToString(randBytes)
}
