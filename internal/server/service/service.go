// Пакет service. Логика сервиса
package service

import (
	"context"

	"github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm"
	"github.com/iurnickita/gophkeeper1/internal/server/model"
	"github.com/iurnickita/gophkeeper1/internal/server/service/config"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
	"go.uber.org/zap"
)

// Service интерфейс сервиса
type Service interface {
	List(ctx context.Context, userID int) ([]string, error)
	Read(ctx context.Context, userID int, unitName string) (model.Unit, error)
	Write(ctx context.Context, unit model.Unit) error
	Delete(ctx context.Context, userID int, unitName string) error
}

// service реализация сервиса
type service struct {
	cfg     config.Config
	store   store.Store
	crypter aesgcm.Crypter
	zaplog  *zap.Logger
}

// List возвращает список доступных данных
func (s service) List(ctx context.Context, userID int) ([]string, error) {
	return s.store.List(ctx, userID)
}

// Read читает единицу данных
func (s service) Read(ctx context.Context, userID int, unitName string) (model.Unit, error) {
	s.zaplog.Sugar().Debug("inbound unitname")
	s.zaplog.Sugar().Debug(unitName)

	// Чтение
	unit, err := s.store.Read(ctx, userID, unitName)
	if err != nil {
		s.zaplog.Error(err.Error())
		return model.Unit{}, err
	}
	s.zaplog.Sugar().Debug("read unit")
	s.zaplog.Sugar().Debug(unit)

	// Дешифрование
	decrUnit, err := s.crypter.UnitDecrypt(unit)
	if err != nil {
		return model.Unit{}, err
	}
	s.zaplog.Sugar().Debug("decrypted unit")
	s.zaplog.Sugar().Debug(decrUnit)

	return decrUnit, nil
}

// Write записывает новую единицу данных
func (s service) Write(ctx context.Context, unit model.Unit) error {
	s.zaplog.Sugar().Debug("inbound unit")
	s.zaplog.Sugar().Debug(unit)

	// Шифрование
	encrUnit, err := s.crypter.UnitEncrypt(unit)
	if err != nil {
		return err
	}
	s.zaplog.Sugar().Debug("encrypted unit")
	s.zaplog.Sugar().Debug(encrUnit)

	// Запись
	err = s.store.Write(ctx, encrUnit)
	if err != nil {
		return err
	}
	return nil
}

// Delete удаляет единицу данных
func (s service) Delete(ctx context.Context, userID int, unitName string) error {
	s.zaplog.Sugar().Debug("inbound unitname")
	s.zaplog.Sugar().Debug(unitName)

	// Удаление
	err := s.store.Delete(ctx, userID, unitName)
	if err != nil {
		s.zaplog.Error(err.Error())
		return err
	}

	return nil
}

// NewService создает объект сервиса
func NewService(cfg config.Config, store store.Store, crypter aesgcm.Crypter, zaplog *zap.Logger) (Service, error) {
	service := service{
		cfg:     cfg,
		store:   store,
		crypter: crypter,
		zaplog:  zaplog}

	return &service, nil
}
