package utils

import (
	"fmt"
	"github.com/vuuvv/errors"
)

const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

var byteEdge = []int{0, 1 << 6, 1 << 12, 1 << 18, 1 << 24, 1 << 30}
var maxInt = (1 << 24) - 1
var decodeArr [128]int

func init() {
	for i, c := range encodeURL {
		decodeArr[c] = i
	}
}

func EncodeIntToBase64Like(value int) (string, error) {
	if value > maxInt {
		return "", errors.Errorf("超出生成序列范围，最大值为%d, 输入的值为%d", maxInt, value)
	}
	bytes := make([]byte, 0, 4)
	for i := 0; i < 4; i++ {
		if value >= byteEdge[i] {
			index := (value >> (i * 6)) & 0x3f
			bytes = append(bytes, encodeURL[index])
		} else {
			bytes = append(bytes, encodeURL[0])
		}
	}

	return string(bytes), nil
}

func DecodeBase64LikeToInt(value string) (int, error) {
	if len(value) > 4 {
		return 0, errors.New("超出解析范围")
	}
	var result int
	for i, c := range value {
		v := decodeArr[c]
		result = result + v<<(i*6)
		if result > maxInt {
			return 0, errors.New(fmt.Sprintf("超出解析范围，最大值为%d, 输入的值为%d", maxInt, result))
		}
	}
	return result, nil
}
