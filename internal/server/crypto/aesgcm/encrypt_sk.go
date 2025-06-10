package aesgcm

import (
	"encoding/json"
	"errors"
	"sort"
	"time"
)

var (
	ErrNoEncryptKeys = errors.New("no encryption keys")
)

// Промежуточный ключ (хранится в БД)
type encryptSK struct {
	keys []Key
}

type Key struct {
	EncryptSK string    `json:"encrypt_sk"`
	Begin     time.Time `json:"begin"`
}

// GetActual возвращает последний ключ
func (sk encryptSK) GetActual() (Key, error) {
	len := len(sk.keys)
	if len == 0 {
		return Key{}, ErrNoEncryptKeys
	}
	return sk.keys[len-1], nil
}

// GetOld возвращает архивный ключ для чтения старых записей
func (sk encryptSK) GetOld(uploadedat time.Time) (Key, error) {
	// Поиск ключа по дате
	for i := len(sk.keys) - 1; i >= 0; i-- {
		key := sk.keys[i]
		if key.Begin.Before(uploadedat) {
			return key, nil
		}
	}
	return Key{}, ErrNoEncryptKeys
}

// CreateNewKey возвращает новый ключ в формате JSON для дальнейшей записи
func (sk encryptSK) CreateNewKey() (string, error) {
	// Формирование нового ключа
	var key Key
	key.EncryptSK = createNewKey()
	key.Begin = time.Now()
	// Запись в переменную
	sk.keys = append(sk.keys, key)
	// Сериализация
	jsonKey, err := json.Marshal(key)
	if err != nil {
		return "", nil
	}
	return string(jsonKey), nil
}

// NewEncryptSK принимает набор ключей в формате json
func NewEncryptSK(strings []string) (encryptSK, error) {
	var sk encryptSK

	// Десериализация
	for _, string := range strings {
		var key Key
		err := json.Unmarshal([]byte(string), &key)
		if err != nil {
			return encryptSK{}, err
		}
		sk.keys = append(sk.keys, key)
	}

	// Сортировка по дате (по убыванию)
	sort.Slice(sk.keys, func(i, j int) bool {
		return sk.keys[i].Begin.Before(sk.keys[j].Begin)
	})

	return sk, nil
}
