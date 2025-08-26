// Пакет config. Конфигурация с помощью флагов/переменных среды/значений по умолчанию
package config

import (
	cacheConfig "github.com/iurnickita/gophkeeper1/internal/client/cache/config"
	grpcClientConig "github.com/iurnickita/gophkeeper1/internal/client/grpc_client/client/config"
	loggerConfig "github.com/iurnickita/gophkeeper1/internal/client/logger/config"
	serviceConfig "github.com/iurnickita/gophkeeper1/internal/client/service/config"
)

// Config - общая конфигурация
type Config struct {
	Logger     loggerConfig.Config
	GRPCClient grpcClientConig.Config
	Service    serviceConfig.Config
	Cache      cacheConfig.Config
}

// GetConfig собирает конфигурацию сервиса
func GetConfig() Config {
	cfg := Config{}

	// По умолчанию на момент разработки
	cfg.Cache.FileRepo = "cache/"
	cfg.Cache.ValidPeriod = 1
	cfg.Logger.LogLevel = "info"

	return cfg
}
