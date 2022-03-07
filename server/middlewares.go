package server

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/goid"
	"github.com/vuuvv/orca/request"
	"net/http"
	"sync"
	"time"
)

var contexts = sync.Map{} //map[int64]*gin.Context{}
const AccessTokenContextKey = "AccessToken"

func MiddlewareId(ctx *gin.Context) {
	id := goid.Get()
	contexts.Store(id, ctx)
	ctx.Next()
	contexts.Delete(id)
}

func GetContext() *gin.Context {
	id := goid.Get()
	ctx, ok := contexts.Load(id)
	if !ok {
		panic("Context set incorrect, are you use middleware [MiddlewareId]")
	}
	return ctx.(*gin.Context)
}

func GetContextOptional() *gin.Context {
	id := goid.Get()
	ctx, _ := contexts.Load(id)
	return ctx.(*gin.Context)
}

func GetAccessToken() *AccessToken {
	ctx := GetContext()
	val, ok := ctx.Get(AccessTokenContextKey)
	if !ok {
		return nil
	}
	at, ok := val.(*AccessToken)
	if !ok {
		return nil
	}
	return at
}

func MiddlewareJwt(config *Config, authorization Authorization) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		guard := authorization.GetGuard(ctx.Request)
		// 可匿名访问
		if guard.IsAnonymous() {
			ctx.Next()
			return
		}

		accessToken, needRefresh, err := validJwt(ctx, config)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, request.NewError(http.StatusUnauthorized, err.Error()))
			ctx.Abort()
			return
		}
		ctx.Set(AccessTokenContextKey, accessToken)

		if guard.IsGuard() && !authorization.Authorized(accessToken, ctx.Request) {
			ctx.JSON(http.StatusForbidden, request.NewErrorForbidden())
			ctx.Abort()
			return
		}

		// TODO: check permission

		ctx.Next()

		// 重新写入刷新的token
		if needRefresh {
			accessTokenString, refreshTokenString, err := GenTokens(config, &AccessToken{
				UserId:    accessToken.UserId,
				Username:  accessToken.Username,
				OrgId:     accessToken.OrgId,
				OrgPath:   accessToken.OrgPath,
				Roles:     accessToken.Roles,
				RoleNames: accessToken.RoleNames,
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, request.NewError(http.StatusInternalServerError, err.Error()))
				ctx.Abort()
				return
			}
			WriteTokenToHead(ctx, config, accessTokenString, refreshTokenString)
			WriteTokenToCookies(ctx, config, accessTokenString, refreshTokenString)
		}
	}
}

func validJwt(ctx *gin.Context, config *Config) (accessToken *AccessToken, needRefresh bool, err error) {
	accessTokenHead := ctx.Request.Header.Get(config.AccessTokenHead)
	var cookie *http.Cookie
	var refreshToken *RefreshToken
	if accessTokenHead == "" {
		cookie, err = ctx.Request.Cookie(config.AccessTokenHead)
		if errors.Is(err, http.ErrNoCookie) {
			return nil, false, request.NewErrorUnauthorized()
		}
		if err != nil {
			return nil, false, errors.WithStack(err)
		}

		accessToken, err = ParseAccessToken(cookie.Value, config.JwtSecret)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
	} else {
		accessToken, err = ParseAccessTokenHead(accessTokenHead, config.JwtTokenPrefix, config.JwtTokenPrefix)
	}

	if accessToken.VerifyExpiresAt(time.Now(), true) {
		return accessToken, false, nil
	}

	// 检测refresh token
	if accessTokenHead == "" {
		cookie, err = ctx.Request.Cookie(config.RefreshTokenHead)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		refreshToken, err = ParseRefreshToken(cookie.Value, config.JwtSecret)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
	} else {
		refreshTokenHead := ctx.Request.Header.Get(config.RefreshTokenHead)
		refreshToken, err = ParseRefreshTokenHead(refreshTokenHead, config.JwtSecret, config.JwtTokenPrefix)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
	}

	if !refreshToken.VerifyExpiresAt(time.Now(), true) {
		return nil, false, errors.New("登录超时，请重新登录")
	}

	return accessToken, true, nil
}
