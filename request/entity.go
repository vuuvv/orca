package request

import (
	"github.com/vuuvv/errors"
	"net/http"
)

type Request interface {
	GetUrl() string
	SetUrl(url string)
	GetMethod() string
	SetMethod(method string)
	GetRequest() string
	SetRequest(req string)
	GetResponse() string
	SetResponse(resp string)
}

type BaseRequest struct {
	RequestUrl    string `json:"url"`
	RequestMethod string `json:"method"`
	Request       string `json:"request"`
	Response      string `json:"response"`
}

func (this *BaseRequest) GetUrl() string {
	return this.RequestUrl
}

func (this *BaseRequest) SetUrl(url string) {
	this.RequestUrl = url
}

func (this *BaseRequest) GetMethod() string {
	return this.RequestMethod
}

func (this *BaseRequest) SetMethod(method string) {
	this.RequestMethod = method
}

func (this *BaseRequest) GetRequest() string {
	return this.Request
}

func (this *BaseRequest) SetRequest(req string) {
	this.Request = req
}

func (this *BaseRequest) GetResponse() string {
	return this.Response
}

func (this *BaseRequest) SetResponse(resp string) {
	this.Response = resp
}

func DoRequest(req Request) (string, error) {
	var body []byte
	var err error

	switch req.GetMethod() {
	case http.MethodGet:
		body, err = GetRaw(req.GetUrl(), nil)
	case http.MethodPost:
		body, err = PostRaw(req.GetUrl(), req.GetRequest())
	case http.MethodPut:
		body, err = PutRaw(req.GetUrl(), req.GetRequest())
	case http.MethodDelete:
		body, err = DeleteRaw(req.GetUrl(), req.GetRequest())
	}
	req.SetResponse(string(body))
	return req.GetResponse(), errors.WithStack(err)
}
