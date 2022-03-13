package server

import (
	rawErrors "errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/id"
	"strings"
	"time"
)

var ErrInvalidAuthHeader = rawErrors.New("auth header is invalid")

type AccessToken struct {
	Id        int64    `json:"id"`
	UserId    int64    `json:"userId"`
	Username  string   `json:"username"`
	OrgId     int64    `json:"orgId"`
	OrgPath   string   `json:"orgPath"`
	Roles     []int64  `json:"roles"`
	RoleNames []string `json:"roleNames"`
	jwt.RegisteredClaims
}

func (this *AccessToken) IsSuper() bool {
	for _, r := range this.RoleNames {
		if r == "system_manager" {
			return true
		}
	}
	return false
}

type RefreshToken struct {
	UserId int64 `json:"userId"`
	jwt.RegisteredClaims
}

func GenAccessToken(issuer string, liveDuration time.Duration, secret string, accessToken *AccessToken) (token string, err error) {
	if accessToken.Id == 0 {
		accessToken.Id = id.Next()
	}
	accessToken.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(liveDuration)),
		Issuer:    issuer,
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessToken).SignedString([]byte(secret))
	return token, errors.WithStack(err)
}

func GenRefreshToken(issuer string, liveDuration time.Duration, secret string, userId int64) (token string, err error) {
	claim := RefreshToken{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(liveDuration)),
			Issuer:    issuer,
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claim).SignedString([]byte(secret))
	return token, errors.WithStack(err)
}

func ParseAccessToken(tokenString string, secret string) (accessToken *AccessToken, err error) {
	accessToken = &AccessToken{}
	err = ParseToken(tokenString, secret, accessToken)
	return accessToken, errors.WithStack(err)
}

func ParseRefreshToken(tokenString string, secret string) (refreshToken *RefreshToken, err error) {
	refreshToken = &RefreshToken{}
	err = ParseToken(tokenString, secret, refreshToken)
	return refreshToken, errors.WithStack(err)
}

func ParseToken(tokenString string, secret string, token jwt.Claims) (err error) {
	_, err = jwt.ParseWithClaims(tokenString, token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return errors.WithStack(err)
	}
	return
}

func ParseTokenFromHead(head string, secret string, prefix string, token jwt.Claims) (err error) {
	parts := strings.SplitN(head, " ", 2)
	if !(len(parts) == 2 && parts[0] == prefix) {
		return errors.WithStack(ErrInvalidAuthHeader)
	}
	err = ParseToken(parts[0], secret, token)
	return errors.WithStack(err)
}

func ParseAccessTokenHead(head string, secret string, prefix string) (token *AccessToken, err error) {
	token = &AccessToken{}
	err = ParseTokenFromHead(head, secret, prefix, token)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

func ParseRefreshTokenHead(head string, secret string, prefix string) (token *RefreshToken, err error) {
	token = &RefreshToken{}
	err = ParseTokenFromHead(head, secret, prefix, token)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return
}
