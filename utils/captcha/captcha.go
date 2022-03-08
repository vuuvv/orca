package captcha

import (
	"github.com/dchest/captcha"
)

func NewCaptcha(length int, width int, height int) (text string, image *captcha.Image) {
	bytes := captcha.RandomDigits(length)
	strBytes := make([]byte, length)
	for i, b := range bytes {
		strBytes[i] = b + '0'
	}
	image = captcha.NewImage(text, bytes, width, height)
	return string(strBytes), image
}
