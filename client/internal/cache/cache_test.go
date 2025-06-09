package cache

import (
	"testing"

	"github.com/iurnickita/gophkeeper1/client/internal/config"
	"github.com/iurnickita/gophkeeper1/client/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestCache_Token(t *testing.T) {
	const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDkxNTA3MzIsIlVzZXJJRCI6IjEifQ.a5A1xNwdyqWXQ_j5JdghzfpkNjD2ggtxDRKgVDRGDE0"

	var cfg config.Config
	cfg.Cache.FileRepo = ""
	cfg.Cache.ValidPeriod = 1
	cfg.Logger.LogLevel = "debug"

	// Лог
	zaplog, err := logger.NewZapLog(cfg.Logger)
	if err != nil {
		t.Error(err)
	}

	// cache create
	cache, err := NewCache(cfg.Cache, zaplog)
	if err != nil {
		t.Error(err)
	}

	// set token
	cache.SetToken(token)
	// get token
	cacheToken := cache.GetToken()
	require.Equal(t, token, cacheToken)

	// cache save
	err = cache.Close()
	require.NoError(t, err)

	// cache create
	cache, err = NewCache(cfg.Cache, zaplog)
	if err != nil {
		t.Error(err)
	}
	// get token
	cacheToken = cache.GetToken()
	require.Equal(t, token, cacheToken)
}
