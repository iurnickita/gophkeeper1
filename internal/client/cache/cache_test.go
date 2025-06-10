package cache

import (
	"testing"

	"github.com/iurnickita/gophkeeper1/internal/client/config"
	"github.com/iurnickita/gophkeeper1/internal/client/logger"
	"github.com/iurnickita/gophkeeper1/internal/client/model"
	"github.com/stretchr/testify/require"
)

func TestCache_Token(t *testing.T) {
	const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDkxNTA3MzIsIlVzZXJJRCI6IjEifQ.a5A1xNwdyqWXQ_j5JdghzfpkNjD2ggtxDRKgVDRGDE0"

	var cfg config.Config
	cfg.Cache.FileRepo = ""
	cfg.Cache.ValidPeriod = 1
	cfg.Logger.LogLevel = "debug"

	// logger
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

func TestCache_List(t *testing.T) {
	list := []string{"secret1", "secret2"}

	var cfg config.Config
	cfg.Cache.FileRepo = ""
	cfg.Cache.ValidPeriod = 1
	cfg.Logger.LogLevel = "debug"

	// logger
	zaplog, err := logger.NewZapLog(cfg.Logger)
	require.NoError(t, err)
	// cache create
	cache, err := NewCache(cfg.Cache, zaplog)
	require.NoError(t, err)

	// SyncList
	err = cache.SyncList(list)
	require.NoError(t, err)
	// GetList
	respList, err := cache.GetList()
	require.NoError(t, err)
	require.Equal(t, list, respList)

	// cache save
	err = cache.Close()
	require.NoError(t, err)
	// cache create
	cache, err = NewCache(cfg.Cache, zaplog)
	require.NoError(t, err)
	// GetList
	respList, err = cache.GetList()
	require.NoError(t, err)
	require.Equal(t, list, respList)
}

func TestCache_Units(t *testing.T) {
	// В БД остался только secret2
	list := []string{"secret2"}
	// В кэше лежит secret1
	unit1 := model.Unit{
		Name: "secret1",
		Body: model.UnitBody{
			Meta: model.UnitMeta{
				Type: 2,
			},
			Data: []byte("secret1 data"),
		},
	}

	var cfg config.Config
	cfg.Cache.FileRepo = ""
	cfg.Cache.ValidPeriod = 1
	cfg.Logger.LogLevel = "debug"

	// logger
	zaplog, err := logger.NewZapLog(cfg.Logger)
	require.NoError(t, err)
	// cache create
	cache, err := NewCache(cfg.Cache, zaplog)
	require.NoError(t, err)

	// SetUnit
	require.NoError(t, cache.SetUnit(unit1))
	// GetUnit
	respUnit1, err := cache.GetUnit(unit1.Name)
	require.NoError(t, err)
	require.Equal(t, unit1.Name, respUnit1.Name)
	require.Equal(t, unit1.Body.Meta.Type, respUnit1.Body.Meta.Type)
	require.Equal(t, unit1.Body.Data, respUnit1.Body.Data)

	// cache save
	err = cache.Close()
	require.NoError(t, err)
	// cache create
	cache, err = NewCache(cfg.Cache, zaplog)
	require.NoError(t, err)
	// GetUnit
	respUnit1, err = cache.GetUnit(unit1.Name)
	require.NoError(t, err)
	require.Equal(t, unit1.Name, respUnit1.Name)
	require.Equal(t, unit1.Body.Meta.Type, respUnit1.Body.Meta.Type)
	require.Equal(t, unit1.Body.Data, respUnit1.Body.Data)

	// SyncList
	require.NoError(t, cache.SyncList(list))
	// GetUnit
	_, err = cache.GetUnit(unit1.Name)
	require.Error(t, err)
}
