// Пакет service. Логика сервиса
package service

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/iurnickita/gophkeeper1/internal/client/cache"
	grpcclient "github.com/iurnickita/gophkeeper1/internal/client/grpc_client/client"
	"github.com/iurnickita/gophkeeper1/internal/client/model"
	"github.com/iurnickita/gophkeeper1/internal/client/service/config"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrOffline         = errors.New("offline")
	ErrUnitTypeIncorr  = errors.New("unit type incorrect")
	ErrTargetMandatory = errors.New("specify a target is mandatory")
	ErrSourceMandatory = errors.New("specify a source is mandatory")
)

// Service интерфейс сервиса
type Service interface {
	Register(login string, password string) error
	Login(login string, password string) error
	List() ([]string, error)
	Read(unitname string, target string) (string, error)
	Write(unitName string, unitType int, unitValue string, source string) error
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
	if err == nil {
		// Вывод из сервера
		s.cache.SyncList(list)
		return list, nil
	} else {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unavailable:
				// Connection refused - вывод из кэша
				list, err = s.cache.GetList()
				if err != nil {
					return nil, err
				}
				return list, ErrOffline
			default:
				return nil, err
			}
		}
		return nil, err
	}
}

// Read
func (s service) Read(unitname string, target string) (string, error) {
	unit, err := s.client.Read(s.cache.GetToken(), unitname)
	if err == nil {
		// Вывод из сервера
		s.cache.SetUnit(unit)
		return s.unitDataToOutput(unit, target)
	} else {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.Unavailable:
				// Connection refused - вывод из кэша
				unit, err = s.cache.GetUnit(unitname)
				if err != nil {
					return "", err
				}
				return s.unitDataToOutput(unit, target)
			default:
				return "", err
			}
		}
		return "", err
	}
}

func (s service) unitDataToOutput(unit model.Unit, target string) (string, error) {
	var result string
	switch unit.Body.Meta.Type {
	case model.UnitTypeLogin:
		var loginModel model.Login
		err := json.Unmarshal(unit.Body.Data, &loginModel)
		if err != nil {
			return "", err
		}
		var s []string
		s = append(s, loginModel.Login)
		s = append(s, loginModel.Password)
		result = strings.Join(s, " ")
	case model.UnitTypeText:
		result = string(unit.Body.Data)
	case model.UnitTypeBinary:
		// вывод только в файл
		if target == "" {
			return "", ErrTargetMandatory
		}
		file, err := os.Create(target)
		if err != nil {
			return "", err
		}
		defer file.Close()
		_, err = file.Write(unit.Body.Data)
		if err != nil {
			return "", err
		}
		file.Sync()
		return "saved to file " + target, nil
	case model.UnitTypeCard:
		var cardModel model.BankCard
		err := json.Unmarshal(unit.Body.Data, &cardModel)
		if err != nil {
			return "", err
		}
		var s []string
		s = append(s, cardModel.Number)
		s = append(s, cardModel.YearMonthTo)
		s = append(s, cardModel.Name)
		s = append(s, cardModel.Surname)
		s = append(s, cardModel.CVV)
		result = strings.Join(s, " ")
	default:
		result = string(unit.Body.Data)
	}

	// сохранение в файл, если указан путь
	if target != "" {
		file, err := os.Create(target)
		if err != nil {
			return "", err
		}
		defer file.Close()
		_, err = file.WriteString(result)
		if err != nil {
			return "", err
		}
		file.Sync()
		return "saved to file " + target, nil
	} else {
		return result, nil
	}
}

// writeUnit записывает готовые данные в форме model.Unit
func (s service) writeUnit(unit model.Unit) error {
	// Запись на сервер
	s.logger.Sugar().Debug("Unit to write")
	s.logger.Sugar().Debug(unit)
	err := s.client.Write(s.cache.GetToken(), unit)
	if err != nil {
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

// writeLogin формирует model.Unit и записывает
func (s service) writeLogin(unitName string, loginModel model.Login) error {
	loginModelJSON, err := json.Marshal(loginModel)
	if err != nil {
		return err
	}
	unit := model.Unit{Name: unitName, Body: model.UnitBody{Meta: model.UnitMeta{Type: model.UnitTypeLogin}, Data: loginModelJSON}}
	if err := s.writeUnit(unit); err != nil {
		return err
	}
	return nil
}

// writeText формирует model.Unit и записывает
func (s service) writeText(unitName string, text string) error {
	unit := model.Unit{Name: unitName, Body: model.UnitBody{Meta: model.UnitMeta{Type: model.UnitTypeText}, Data: []byte(text)}}
	if err := s.writeUnit(unit); err != nil {
		return err
	}
	return nil
}

// writeBinary формирует model.Unit и записывает
func (s service) writeBinary(unitName string, bytes []byte) error {
	unit := model.Unit{Name: unitName, Body: model.UnitBody{Meta: model.UnitMeta{Type: model.UnitTypeBinary}, Data: bytes}}
	if err := s.writeUnit(unit); err != nil {
		return err
	}
	return nil
}

// writeCard формирует model.Unit и записывает
func (s service) writeCard(unitName string, card model.BankCard) error {
	cardModelJSON, err := json.Marshal(card)
	if err != nil {
		return err
	}
	unit := model.Unit{Name: unitName, Body: model.UnitBody{Meta: model.UnitMeta{Type: model.UnitTypeCard}, Data: cardModelJSON}}
	if err := s.writeUnit(unit); err != nil {
		return err
	}
	return nil
}

// Write
func (s service) Write(unitName string, unitType int, unitValue string, source string) error {
	// Указан путь к файлу - берем оттуда
	var bytes []byte
	var err error
	if source != "" {
		bytes, err = os.ReadFile(source)
		if err != nil {
			return err
		}
		unitValue = string(bytes)
	}

	switch unitType {
	case model.UnitTypeLogin:
		unitValues := strings.Fields(unitValue)
		loginModel := model.Login{
			Login:    unitValues[0],
			Password: unitValues[1],
		}

		if err := s.writeLogin(unitName, loginModel); err != nil {
			return err
		}
	case model.UnitTypeText:
		if err := s.writeText(unitName, unitValue); err != nil {
			return err
		}
	case model.UnitTypeBinary:
		if bytes == nil {
			return ErrSourceMandatory
		}
		if err := s.writeBinary(unitName, bytes); err != nil {
			return err
		}
	case model.UnitTypeCard:
		unitValues := strings.Fields(unitValue)
		cardModel := model.BankCard{
			Number:      unitValues[0],
			YearMonthTo: unitValues[1],
			Name:        unitValues[2],
			Surname:     unitValues[3],
			CVV:         unitValues[4],
		}

		if err := s.writeCard(unitName, cardModel); err != nil {
			return err
		}
	default:
		return ErrUnitTypeIncorr
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
