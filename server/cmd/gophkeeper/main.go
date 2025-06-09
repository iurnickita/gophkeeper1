package main

import (
	"log"

	"github.com/iurnickita/gophkeeper1/server/internal/auth"
	"github.com/iurnickita/gophkeeper1/server/internal/config"
	"github.com/iurnickita/gophkeeper1/server/internal/crypto/aesgcm"
	grpcserver "github.com/iurnickita/gophkeeper1/server/internal/grpc_server/server"
	"github.com/iurnickita/gophkeeper1/server/internal/logger"
	"github.com/iurnickita/gophkeeper1/server/internal/service"
	"github.com/iurnickita/gophkeeper1/server/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	zaplog, err := logger.NewZapLog(cfg.Logger)
	if err != nil {
		return err
	}

	store, err := store.NewStore(cfg.Store)
	if err != nil {
		return err
	}

	auth, err := auth.NewAuth(store)
	if err != nil {
		return err
	}

	crypter, err := aesgcm.NewCrypter(cfg.Crypter, store)
	if err != nil {
		return err
	}

	service, err := service.NewService(cfg.Service, store, crypter, zaplog)
	if err != nil {
		return err
	}

	return grpcserver.Serve(cfg.GRPCServer, auth, service, zaplog)
}
