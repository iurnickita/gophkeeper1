// Пакет config. Конфигурация с помощью флагов/переменных среды/значений по умолчанию
package config

import (
	"flag"
	"os"

	crypterConfig "github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm/config"
	grpcServerConig "github.com/iurnickita/gophkeeper1/internal/server/grpc_server/server/config"
	loggerConfig "github.com/iurnickita/gophkeeper1/internal/server/logger/config"
	serviceConfig "github.com/iurnickita/gophkeeper1/internal/server/service/config"
	storeConfig "github.com/iurnickita/gophkeeper1/internal/server/store/config"
)

// Config - общая конфигурация
type Config struct {
	GRPCServer grpcServerConig.Config
	Service    serviceConfig.Config
	Store      storeConfig.Config
	Crypter    crypterConfig.Config
	Logger     loggerConfig.Config
}

// GetConfig собирает конфигурацию сервиса
func GetConfig() Config {
	cfg := Config{}

	// Флаги
	flag.StringVar(&cfg.Store.DBDsn, "d", "", "database dsn")
	flag.StringVar(&cfg.Logger.LogLevel, "l", "info", "log level")

	// Переменные окружения
	if envdsn := os.Getenv("DATABASE_URI"); envdsn != "" {
		cfg.Store.DBDsn = envdsn
	}
	if envlevel := os.Getenv("LOG_LEVEL"); envlevel != "" {
		cfg.Logger.LogLevel = envlevel
	}

	// По умолчанию на момент разработки
	cfg.Store.DBDsn = "host=localhost user=bob password=bob dbname=gophkeeper sslmode=disable"
	cfg.Crypter.MasterSK = "cb459063d4bbbd4ce04a7c5b6e8121e7933630bada8fcb3abc20f6ca0aba3793"
	cfg.Crypter.NewSKIntervalD = 30

	return cfg
}
