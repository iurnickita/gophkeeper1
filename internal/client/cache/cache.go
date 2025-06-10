// Пакет cache. Кэширование
package cache

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/iurnickita/gophkeeper1/internal/client/cache/config"
	"github.com/iurnickita/gophkeeper1/internal/client/model"
	"go.uber.org/zap"
)

// Cache - интерфейс кэша
type Cache interface {
	GetList() ([]string, error)
	SyncList(serverList []string) error
	GetUnit(unitName string) (model.Unit, error)
	SetUnit(unit model.Unit) error
	DeleteUnit(unitName string) error
	GetToken() string
	SetToken(token string)
	Close() error
}

var (
	ErrNotFound = errors.New("data not found")
)

const (
	ListFileName  = "dataList.txt"
	UnitsFileName = "dataUnits.json"
	TokenFileName = "token.txt"
)

type cache struct {
	cfg    config.Config
	list   list
	units  units
	token  token
	logger *zap.Logger
}

type list struct {
	list []string
	mux  *sync.Mutex
	chg  bool
	file *os.File
}

type units struct {
	units map[string]model.Unit
	mux   *sync.Mutex
	chg   bool
	file  *os.File
}

type token struct {
	token string
	chg   bool
	file  *os.File
}

// GetList возвращает список доступных данных
func (c *cache) GetList() ([]string, error) {
	return c.list.list, nil
}

// SyncList
func (c *cache) SyncList(serverList []string) error {
	c.list.mux.Lock()
	defer c.list.mux.Unlock()

	c.list.list = serverList
	c.list.chg = true
	return nil
}

// GetUnit
func (c *cache) GetUnit(unitName string) (model.Unit, error) {
	c.units.mux.Lock()
	defer c.units.mux.Unlock()
	c.list.mux.Lock()
	defer c.list.mux.Unlock()

	// Поиск по списку
	idx := slices.Index(c.list.list, unitName)
	if idx == -1 {
		return model.Unit{}, ErrNotFound
	}
	// Поиск данных
	unit, ok := c.units.units[unitName]
	if !ok {
		return model.Unit{}, ErrNotFound
	}
	// Проверка даты действия
	if unit.Body.Meta.ValidUntil.IsZero() {
		return model.Unit{}, ErrNotFound
	}
	if unit.Body.Meta.ValidUntil.Before(time.Now()) {
		return model.Unit{}, ErrNotFound
	}

	return unit, nil
}

// SetUnit
func (c *cache) SetUnit(unit model.Unit) error {
	c.units.mux.Lock()
	defer c.units.mux.Unlock()
	c.list.mux.Lock()
	defer c.list.mux.Unlock()

	c.logger.Sugar().Debug("cache.SetUnit unit:")
	c.logger.Sugar().Debug(unit)

	// Установка даты действия
	unit.Body.Meta.ValidUntil = time.Now().AddDate(0, 0, c.cfg.ValidPeriod)
	// Запись
	c.list.list = append(c.list.list, unit.Name)
	c.list.chg = true
	c.units.units[unit.Name] = unit
	c.units.chg = true
	return nil
}

// DeleteUnit
func (c *cache) DeleteUnit(unitName string) error {
	c.units.mux.Lock()
	defer c.units.mux.Unlock()

	// Удаление из списка
	idx := slices.Index(c.list.list, unitName)
	if idx == -1 {
		return ErrNotFound
	}
	c.list.list = slices.Delete(c.list.list, idx, idx+1)
	return nil
}

// GetToken
func (c *cache) GetToken() string {
	return c.token.token
}

// SetToken
func (c *cache) SetToken(token string) {
	c.token.token = token
	c.token.chg = true
}

// Close сохраняет данные и закрывает файлы
func (c *cache) Close() error {
	err := c.saveList()
	if err != nil {
		return err
	}
	err = c.saveUnits()
	if err != nil {
		return err
	}
	err = c.saveToken()
	if err != nil {
		return err
	}

	c.list.file.Close()
	c.units.file.Close()
	c.token.file.Close()

	return nil
}

// saveList сохраняет список в файл
func (c *cache) saveList() error {
	// были изменения
	if !c.list.chg {
		return nil
	}

	// перезапись файла
	err := c.list.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = c.list.file.Seek(0, 0)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(c.list.file)
	for _, unitName := range c.list.list {
		// записываем в буфер
		if _, err := writer.Write([]byte(unitName)); err != nil {
			return err
		}
		// добавляем перенос строки
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	// записываем буфер в файл
	writer.Flush()
	return nil
}

// saveUnits сохраняет полезные данные в файл
func (c *cache) saveUnits() error {
	// были изменения
	if !c.units.chg {
		return nil
	}

	// перезапись файла
	err := c.units.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = c.units.file.Seek(0, 0)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(c.units.file)
	for _, unit := range c.units.units {
		// сериализация
		jsonLine, err := json.Marshal(unit)
		if err != nil {
			return err
		}
		// записываем в буфер
		if _, err := writer.Write(jsonLine); err != nil {
			return err
		}
		// добавляем перенос строки
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	// записываем буфер в файл
	writer.Flush()
	return nil
}

// saveToken сохраняет токен в файл
func (c *cache) saveToken() error {
	// были изменения
	if !c.token.chg {
		return nil
	}

	// перезапись файла
	err := c.token.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = c.token.file.Seek(0, 0)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(c.token.file)
	// записываем в буфер
	_, err = writer.WriteString(c.token.token)
	if err != nil {
		return err
	}
	// записываем буфер в файл
	writer.Flush()
	return nil
}

// init подготавливает данные из файлов
func (c *cache) init() error {
	// Чтение списка
	list, err := c.initList()
	if err != nil {
		return err
	}
	// Чтение полезных данных
	units, err := c.initUnits()
	if err != nil {
		return err
	}
	// Чтение токена
	token, err := c.initToken()
	if err != nil {
		return err
	}

	c.list = list
	c.units = units
	c.token = token
	return nil
}

// initList чтение списка
func (c *cache) initList() (list, error) {
	var l list

	file, err := os.OpenFile(c.cfg.FileRepo+ListFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return list{}, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l.list = append(l.list, string(scanner.Bytes()))
	}
	l.mux = &sync.Mutex{}
	l.file = file
	return l, nil
}

// initUnits чтение полезных данных
func (c *cache) initUnits() (units, error) {
	var u units
	u.units = make(map[string]model.Unit)

	file, err := os.OpenFile(c.cfg.FileRepo+UnitsFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return units{}, err
	}
	scanner := bufio.NewScanner(file)
	var unit model.Unit
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &unit); err == nil {
			u.units[unit.Name] = unit
		}
	}
	u.mux = &sync.Mutex{}
	u.file = file
	return u, nil
}

// initToken чтение токена
func (c *cache) initToken() (token, error) {
	file, err := os.OpenFile(c.cfg.FileRepo+TokenFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return token{}, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	tokenString := scanner.Text()

	var token token
	token.token = tokenString
	token.file = file
	return token, nil
}

// NewCache создает объект кэша
func NewCache(cfg config.Config, logger *zap.Logger) (Cache, error) {
	var cache cache
	cache.cfg = cfg
	cache.logger = logger
	cache.init()

	return &cache, nil
}
