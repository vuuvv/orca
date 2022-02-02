package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenAccessToken(t *testing.T) {
	secret := "test"
	tokenString, err := GenAccessToken("orca", 5*time.Minute, secret, 123, "admin", nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(tokenString)
	token, err := ParseAccessToken(tokenString, secret)
	if err != nil {
		t.Error(err)
	}

	t.Log(token.ExpiresAt)
	assert.Equal(t, "orca", token.Issuer)
	assert.Equal(t, int64(123), token.UserId)
	assert.Equal(t, "admin", token.Username)
}

func TestGenRefreshToken(t *testing.T) {
	secret := "test"
	tokenString, err := GenRefreshToken("orca", 5*time.Minute, secret, 123)
	if err != nil {
		t.Error(err)
	}
	t.Log(tokenString)
	token, err := ParseRefreshToken(tokenString, secret)
	if err != nil {
		t.Error(err)
	}

	t.Log(token.ExpiresAt)
	assert.Equal(t, "orca", token.Issuer)
	assert.Equal(t, int64(123), token.UserId)
}
