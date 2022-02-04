package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/vuuvv/orca/utils"
	"html/template"
	"net"
	"sync"
	"unsafe"
)

type nodeType uint8
type HandlerFunc func(*gin.Context)
type HandlersChain []HandlerFunc

type engine struct {
	gin.RouterGroup

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool

	// DEPRECATED: USE `TrustedPlatform` WITH VALUE `gin.GoogleAppEngine` INSTEAD
	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool

	// If enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// If true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool

	// List of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	RemoteIPHeaders []string

	// If set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	TrustedPlatform string

	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64

	delims           render.Delims
	secureJSONPrefix string
	HTMLRender       render.HTMLRender
	FuncMap          template.FuncMap
	allNoRoute       HandlersChain
	allNoMethod      HandlersChain
	noRoute          HandlersChain
	noMethod         HandlersChain
	pool             sync.Pool
	trees            methodTrees
	maxParams        uint16
	maxSections      uint16
	trustedProxies   []string
	trustedCIDRs     []*net.IPNet
}

type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node // child nodes, at most 1 :param style node at the end of the array
	handlers  HandlersChain
	fullPath  string
}

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

type Route struct {
	Group      string
	Name       string
	Method     string
	Path       string
	Handler    string
	Permission Guard
}

func (this *Route) WithName(name string) *Route {
	this.Name = name
	return this
}

func (this *Route) Anonymous() *Route {
	this.Permission = GuardAnonymous
	return this
}

func (this *Route) Login() *Route {
	this.Permission = GuardLogin
	return this
}

func Routes(gin *gin.Engine) []*Route {
	e := (*engine)(unsafe.Pointer(gin))
	var ret []*Route
	for _, n := range e.trees {
		ret = append(ret, walk(n.method, n.root)...)
	}
	return ret
}

func walk(method string, n *node) []*Route {
	var ret []*Route
	if len(n.handlers) > 0 {
		ret = append(ret, &Route{
			Method:  method,
			Path:    n.fullPath,
			Handler: utils.FunctionName(n.handlers[len(n.handlers)-1]),
		})
	}

	if len(n.children) > 0 {
		for _, child := range n.children {
			cRet := walk(method, child)
			ret = append(ret, cRet...)
		}
	}

	return ret
}
