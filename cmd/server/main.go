package main

import (
	"fmt"
	"log"

	"github.com/iurnickita/gophkeeper1/internal/server/auth"
	"github.com/iurnickita/gophkeeper1/internal/server/config"
	"github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm"
	grpcserver "github.com/iurnickita/gophkeeper1/internal/server/grpc_server/server"
	"github.com/iurnickita/gophkeeper1/internal/server/logger"
	"github.com/iurnickita/gophkeeper1/internal/server/service"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
)

// -ldflags
var (
	buildVersion string
	buildDate    string
	bulidCommit  string
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Флаги сборки (флаги линковщика)
	fmt.Printf("buildVersion: %s\n", fillEmptyFlag(buildVersion))
	fmt.Printf("buildDate: %s\n", fillEmptyFlag(buildDate))
	fmt.Printf("bulidCommit: %s\n", fillEmptyFlag(bulidCommit))

	// Config
	cfg := config.GetConfig()

	// Лог
	zaplog, err := logger.NewZapLog(cfg.Logger)
	if err != nil {
		return err
	}

	// Хранилище
	store, err := store.NewStore(cfg.Store)
	if err != nil {
		return err
	}

	// Аутентификация
	auth, err := auth.NewAuth(store)
	if err != nil {
		return err
	}

	// Шифровальщик
	crypter, err := aesgcm.NewCrypter(cfg.Crypter, store)
	if err != nil {
		return err
	}

	// Сервис
	service, err := service.NewService(cfg.Service, store, crypter, zaplog)
	if err != nil {
		return err
	}

	// Хендлер
	return grpcserver.Serve(cfg.GRPCServer, auth, service, zaplog)
}

func fillEmptyFlag(s string) string {
	if s == "" {
		s = "N/A"
	}
	return s
}
