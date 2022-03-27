package secure

import (
	"github.com/meehow/securebytes"
	"github.com/vuuvv/errors"
)

func NewSecure(token string) *securebytes.SecureBytes {
	return securebytes.New([]byte(token), securebytes.ASN1Serializer{})
}

var secure *securebytes.SecureBytes

func GetSecure() *securebytes.SecureBytes {
	return secure
}

func SetSecure(s *securebytes.SecureBytes) {
	secure = s
}

func Encrypt(input string) (string, error) {
	if secure == nil {
		return "", errors.New("加密解密未初始话")
	}
	output, err := secure.RawEncryptToBase64([]byte(input))
	return output, errors.WithStack(err)
}

func Decrypt(input string) (string, error) {
	if secure == nil {
		return "", errors.New("加密解密未初始话")
	}
	output, err := secure.RawDecryptBase64(input)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(output), nil
}
