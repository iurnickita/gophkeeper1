package main

import (
	"fmt"
	"log"

	"github.com/iurnickita/gophkeeper1/internal/client/cache"
	"github.com/iurnickita/gophkeeper1/internal/client/cli"
	"github.com/iurnickita/gophkeeper1/internal/client/config"
	grpcclient "github.com/iurnickita/gophkeeper1/internal/client/grpc_client/client"
	"github.com/iurnickita/gophkeeper1/internal/client/logger"
	"github.com/iurnickita/gophkeeper1/internal/client/service"
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

func fillEmptyFlag(s string) string {
	if s == "" {
		s = "N/A"
	}
	return s
}
