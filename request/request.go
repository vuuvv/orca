package request

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/vuuvv/errors"
	"net"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

var client *resty.Client

type doRequest func(url string) (*resty.Response, error)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *Response) IsSuccess() bool {
	return r.Code == 0
}

func (r *Response) IsError() bool {
	return !r.IsSuccess()
}

func Get(url string, params map[string]string, resp interface{}) ([]byte, error) {
	return request(url, resp, get(params))
}

func GetRaw(url string, params map[string]string) ([]byte, error) {
	return requestRaw(url, get(params))
}

func Post(url string, data interface{}, resp interface{}) ([]byte, error) {
	return request(url, resp, dataHandler(http.MethodPost, data))
}

func PostRaw(url string, data interface{}) ([]byte, error) {
	return requestRaw(url, dataHandler(http.MethodPost, data))
}

func Put(url string, data interface{}, resp interface{}) ([]byte, error) {
	return request(url, resp, dataHandler(http.MethodPut, data))
}

func PutRaw(url string, data interface{}) ([]byte, error) {
	return requestRaw(url, dataHandler(http.MethodPut, data))
}

func Delete(url string, data interface{}, resp interface{}) ([]byte, error) {
	return request(url, resp, dataHandler(http.MethodDelete, data))
}

func DeleteRaw(url string, data interface{}) ([]byte, error) {
	return requestRaw(url, dataHandler(http.MethodDelete, data))
}

func get(params map[string]string) func(url string) (*resty.Response, error) {
	return func(url string) (*resty.Response, error) {
		return GetClient().R().
			SetQueryParams(params).
			SetHeader("Accept", "application/json").
			Get(url)
	}
}

func dataHandler(typ string, data interface{}) func(url string) (*resty.Response, error) {
	return func(url string) (*resty.Response, error) {
		if data != nil {
			kind := reflect.Indirect(reflect.ValueOf(data)).Type().Kind()

			if kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice {
				bytes, err := jsoniter.Marshal(data)
				if err != nil {
					return nil, err
				}
				data = bytes
			}
		}

		req := GetClient().R().
			SetBody(data).
			SetHeader("Accept", "application/json")
		switch typ {
		case http.MethodPost:
			return req.Post(url)
		case http.MethodPut:
			return req.Put(url)
		case http.MethodDelete:
			return req.Delete(url)
		}
		return nil, errors.New(fmt.Sprintf("不支持的http method [%s]", typ))
	}
}

func request(url string, resp interface{}, handler doRequest) ([]byte, error) {
	bytes, err := requestRaw(url, handler)
	if err != nil {
		return bytes, errors.WithStack(err)
	}
	ret := Response{Data: resp}
	err = jsoniter.Unmarshal(bytes, &ret)
	if ret.IsError() {
		return bytes, NewError(ret.Code, ret.Message).WithStack()
	}
	return bytes, err
}

func requestRaw(url string, handler doRequest) ([]byte, error) {
	resp, err := handler(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.Body(), nil
}

func GetClient() *resty.Client {
	if client == nil {
		client = resty.New()
		client.SetTransport(createTransport(nil))
	}
	return client
}

func createTransport(localAddr net.Addr) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 600 * time.Second,
		DualStack: true,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}
