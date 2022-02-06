package server

import "net/http"

type Guard string

func (this Guard) IsAnonymous() bool {
	return this == GuardAnonymous
}

func (this Guard) IsLogin() bool {
	return this == GuardLogin
}

func (this Guard) IsGuard() bool {
	return this == GuardAuthorization
}

const (
	GuardAnonymous     Guard = "anonymous"     // GuardAnonymous 匿名访问
	GuardLogin         Guard = "login"         // GuardLogin 需登录访问
	GuardAuthorization Guard = "authorization" // GuardPermission 需有权限才可访问
)

type Authorization interface {
	GetGuard(request *http.Request) Guard
	Authorized(accessToken *AccessToken, request *http.Request) bool
	Refresh() error
}

type SimpleAuthorization struct {
	AnonymousRoutes map[string]bool
}

func (d SimpleAuthorization) GetGuard(request *http.Request) Guard {
	if d.AnonymousRoutes != nil {
		if _, ok := d.AnonymousRoutes[request.RequestURI]; ok {
			return GuardAnonymous
		}
	}
	return GuardLogin
}

func (d SimpleAuthorization) Authorized(accessToken *AccessToken, request *http.Request) bool {
	return true
}

func (d SimpleAuthorization) Refresh() error {
	return nil
}
