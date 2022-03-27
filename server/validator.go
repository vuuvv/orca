package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/govalidator"
	"github.com/vuuvv/orca/request"
	"io"
	"strconv"
)

type Validator struct {
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *Validator) ValidateStruct(obj interface{}) error {
	_, err := govalidator.ValidateStruct(obj)
	return err
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://godoc.org/gopkg.in/go-playground/validator.v8
func (v *Validator) Engine() interface{} {
	return nil
}

func Parse(ctx *gin.Context, val interface{}) (err error) {
	if val == nil {
		return errors.New("can not parse nil pointer")
	}
	decoder := jsoniter.NewDecoder(ctx.Request.Body)
	if err = decoder.Decode(val); err != nil {
		if errors.Is(err, io.EOF) {
			return request.ErrorNoArgument("")
		}
		return request.ErrorBadRequest(err.Error())
	}
	return nil
}

func ParseForm(ctx *gin.Context, val interface{}) (err error) {
	if binding.Validator == nil {
		return errors.New("binding validator未初始化")
	}
	if err = Parse(ctx, val); err != nil {
		return err
	}
	return errors.WithStack(binding.Validator.ValidateStruct(val))
}

func MustParseForm(ctx *gin.Context, val interface{}) {
	err := ParseForm(ctx, val)
	if err != nil {
		panic(err)
	}
}

func ParseIds(ctx *gin.Context) (ids []int64, err error) {
	err = Parse(ctx, &ids)
	return
}

func ParseQueryInt(ctx *gin.Context, key string, required bool) (value int, err error) {
	if str, ok := ctx.GetQuery(key); ok {
		if value, err = strconv.Atoi(str); err != nil {
			return 0, request.ErrorBadRequest("解析int错误：%s", err.Error())
		}
	} else {
		if required {
			return 0, request.ErrorNoArgument(key)
		}
	}
	return value, err
}

func ParseQueryInt64(ctx *gin.Context, key string, required bool) (value int64, err error) {
	if str, ok := ctx.GetQuery(key); ok {
		if value, err = strconv.ParseInt(str, 10, 64); err != nil {
			return 0, request.ErrorBadRequest("解析int64错误：%s", err.Error())
		}
	} else {
		if required {
			return 0, request.ErrorNoArgument(key)
		}
	}

	return value, err
}
