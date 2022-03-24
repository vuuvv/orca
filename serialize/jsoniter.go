package serialize

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
	"strconv"
	"time"
	"unsafe"
)

func InitializeJsoniter() {
	RegisterEncodeInt64()
	RegisterDecodeInt64()
	RegisterDecodeInt()
	RegisterEncodeTime()
	RegisterDecodeTime()
	RegisterEncodeError()
}

func MustJsonStringify(value interface{}) string {
	str, err := JsonStringify(value)
	if err != nil {
		panic(err)
	}
	return str
}

func JsonStringify(value interface{}) (string, error) {
	switch val := value.(type) {
	case string:
		return val, nil
	case []byte:
		return string(val), nil
	}
	bytes, err := jsoniter.Marshal(value)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(bytes), nil
}

func MustJsonStringifyBytes(value interface{}) []byte {
	bytes, err := JsonStringifyBytes(value)
	if err != nil {
		panic(err)
	}
	return bytes
}

func JsonStringifyBytes(value interface{}) ([]byte, error) {
	switch val := value.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	}
	bytes, err := jsoniter.Marshal(value)
	return bytes, errors.WithStack(err)
}

func MustJsonParse[T any](json string) *T {
	return MustJsonParseBytes[T]([]byte(json))
}

func JsonParse[T any](json string) (*T, error) {
	return JsonParseBytes[T]([]byte(json))
}

func MustJsonParsePrimitive[T any](json string) T {
	return MustJsonParseBytesPrimitive[T]([]byte(json))
}

func JsonParsePrimitive[T any](json string) (T, error) {
	return JsonParseBytesPrimitive[T]([]byte(json))
}

func MustJsonParseBytes[T any](json []byte) *T {
	val, err := JsonParseBytes[T](json)
	if err != nil {
		panic(err)
	}
	return val
}

func JsonParseBytes[T any](json []byte) (*T, error) {
	var ret T
	err := jsoniter.Unmarshal(json, &ret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ret, nil
}

func MustJsonParseBytesPrimitive[T any](json []byte) T {
	val, err := JsonParseBytesPrimitive[T](json)
	if err != nil {
		panic(err)
	}
	return val
}

func JsonParseBytesPrimitive[T any](json []byte) (T, error) {
	var ret T
	err := jsoniter.Unmarshal(json, &ret)
	if err != nil {
		return ret, errors.WithStack(err)
	}
	return ret, nil
}

func RegisterEncodeInt64() {
	jsoniter.RegisterTypeEncoderFunc("int64", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		n := *((*int64)(ptr))
		stream.WriteString(strconv.FormatInt(n, 10))
	}, nil)
}

func RegisterDecodeInt64() {
	jsoniter.RegisterTypeDecoderFunc("int64", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		any := iter.Read()
		if any == nil {
			*((*int64)(ptr)) = 0
			return
		}
		if f64, ok := any.(float64); ok {
			*((*int64)(ptr)) = int64(f64)
		} else if i, ok := any.(int); ok {
			*((*int64)(ptr)) = int64(i)
		} else if str, ok := any.(string); ok {
			if str == "" {
				*((*int64)(ptr)) = 0
			} else {
				t, err := strconv.ParseInt(str, 10, 64)
				if err != nil {
					iter.Error = err
					return
				}
				*((*int64)(ptr)) = t
			}
		} else {
			*((*int64)(ptr)) = any.(int64)
		}
	})
}

func RegisterDecodeInt() {
	jsoniter.RegisterTypeDecoderFunc("int", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		any := iter.Read()
		if any == nil {
			*((*int)(ptr)) = 0
			return
		}
		if f64, ok := any.(float64); ok {
			*((*int)(ptr)) = int(f64)
		} else if i, ok := any.(int); ok {
			*((*int)(ptr)) = i
		} else if str, ok := any.(string); ok {
			if str == "" {
				*((*int)(ptr)) = 0
			} else {
				t, err := strconv.Atoi(str)
				if err != nil {
					iter.Error = err
					return
				}
				*((*int)(ptr)) = t
			}
		} else {
			*((*int)(ptr)) = any.(int)
		}
	})
}

func RegisterEncodeError() {
	jsoniter.RegisterTypeEncoderFunc("error", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		n := *((*error)(ptr))
		if e, ok := n.(govalidator.Error); ok {
			stream.WriteVal(e)
			return
		}
		if n == nil {
			stream.WriteNil()
		} else {
			stream.WriteString(n.Error())
		}
	}, nil)
}

func RegisterEncodeTime() {
	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		stream.WriteString(t.Format("2006-01-02 15:04:05"))
	}, nil)
}

func RegisterDecodeTime() {
	jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", iter.ReadString(), time.Local)
		if err != nil {
			iter.Error = err
			return
		}
		*((*time.Time)(ptr)) = t
	})
}
