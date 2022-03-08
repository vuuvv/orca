package captcha

import (
	"fmt"
	"github.com/dchest/captcha"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	bytes := captcha.RandomDigits(4)
	fmt.Println(bytes)
}
