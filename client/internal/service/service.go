// Пакет service. Логика сервиса
package service

import (
	"errors"

	"github.com/iurnickita/gophkeeper1/client/internal/cache"
	grpcclient "github.com/iurnickita/gophkeeper1/client/internal/grpc_client/client"
	"github.com/iurnickita/gophkeeper1/client/internal/model"
	"github.com/iurnickita/gophkeeper1/client/internal/service/config"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrOffline = errors.New("offline")
)

// Service интерфейс сервиса
type Service interface {
	Register(login string, password string) error
	Login(login string, password string) error
	List() ([]string, error)
	Read(unitname string) (model.Unit, error)
	Write(unit model.Unit) error
	Delete(unitname string) error
	Close()
}

type service struct {
	cfg    config.Config
	client grpcclient.Client
	cache  cache.Cache
	logger *zap.Logger
}

// Register
func (s service) Register(login string, password string) error {
	token, err := s.client.Register(login, password)
	if err != nil {
		return err
	}
	s.logger.Sugar().Debugf("register returns token: %s", token)
	s.cache.SetToken(token)
	return nil
}

// Login
func (s service) Login(login string, password string) error {
	token, err := s.client.Authenticate(login, password)
	if err != nil {
		return err
	}
	s.logger.Sugar().Debugf("authenticate returns token: %s", token)
	s.cache.SetToken(token)
	return nil
}

// List
func (s service) List() ([]string, error) {
	list, err := s.client.List(s.cache.GetToken())
	switch err {
	case nil:
		// Вывод из сервера
		s.cache.SyncList(list)
		return list, nil
	default:
		// Вывод из кэша
		// тут надо отличать ошибку соединения от остальных
		// case "connection_refused":
		list, err = s.cache.GetList()
		if err != nil {
			return nil, err
		}
		return list, ErrOffline
	}
}

// Read
func (s service) Read(unitname string) (model.Unit, error) {
	unit, err := s.client.Read(s.cache.GetToken(), unitname)
	if err == nil {
		// Вывод из сервера
		s.cache.SetUnit(unit)
		return unit, nil
	} else {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unavailable:
				// Connection refused - вывод из кэша
				unit, err = s.cache.GetUnit(unitname)
				if err != nil {
					return model.Unit{}, nil
				}
				return unit, ErrOffline
			default:
				return model.Unit{}, err
			}
		}
		return model.Unit{}, err
	}
}

// Write
func (s service) Write(unit model.Unit) error {
	// Запись на сервер
	s.logger.Sugar().Debug("Unit to write")
	s.logger.Sugar().Debug(unit)
	err := s.client.Write(s.cache.GetToken(), unit)
	if err != nil {
		// Это надо убрать
		// А на стороне сервера сделать перезапись
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.AlreadyExists:
				return err
			default:
				return err
			}
		} else {
			return err
		}
	}
	// Кэширование
	err = s.cache.SetUnit(unit)
	if err != nil {
		return err
	}
	return nil
}

// Delete
func (s service) Delete(unitname string) error {
	// Удаление с сервера
	err := s.client.Delete(s.cache.GetToken(), unitname)
	if err != nil {
		return err
	}
	// Удаление кэша
	err = s.cache.DeleteUnit(unitname)
	if err != nil {
		return err
	}
	return nil
}

// Close
func (s service) Close() {
	s.client.Close()
	s.cache.Close()
}

// NewService создает сервис
func NewService(cfg config.Config, client grpcclient.Client, cache cache.Cache, logger *zap.Logger) (Service, error) {
	return service{cfg: cfg, client: client, cache: cache, logger: logger}, nil
}
