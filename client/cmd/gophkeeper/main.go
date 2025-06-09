package main

import (
	"log"

	"github.com/iurnickita/gophkeeper1/client/internal/cache"
	"github.com/iurnickita/gophkeeper1/client/internal/cli"
	"github.com/iurnickita/gophkeeper1/client/internal/config"
	grpcclient "github.com/iurnickita/gophkeeper1/client/internal/grpc_client/client"
	"github.com/iurnickita/gophkeeper1/client/internal/logger"
	"github.com/iurnickita/gophkeeper1/client/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	// Лог
	zaplog, err := logger.NewZapLog(cfg.Logger)
	if err != nil {
		return err
	}

	// Клиент
	client, err := grpcclient.NewClient(cfg.GRPCClient)
	if err != nil {
		return err
	}

	// Кэш
	cache, err := cache.NewCache(cfg.Cache, zaplog)
	if err != nil {
		return err
	}

	// Логика
	service, err := service.NewService(cfg.Service, client, cache, zaplog)
	if err != nil {
		return err
	}

	// Пользовательский интерфейс
	cli.Execute(service)

	// Завершение работы
	service.Close()
	return nil
}
