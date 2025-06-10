package aesgcm

import (
	"context"
	"time"

	"github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm/config"
	"github.com/iurnickita/gophkeeper1/internal/server/model"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
)

type Crypter interface {
	UnitEncrypt(unit model.Unit) (model.Unit, error)
	UnitDecrypt(unit model.Unit) (model.Unit, error)
}

type crypter struct {
	cfg       config.Config
	encryptSK encryptSK
}

func (c crypter) UnitEncrypt(unit model.Unit) (model.Unit, error) {
	// Шифрование уникальным ключом
	unitSK := createNewKey()
	encrString, err := encrypt(string(unit.Data), unitSK)
	if err != nil {
		return model.Unit{}, err
	}
	// Запись зашифрованных данных
	unit.Data = []byte(encrString)
	// Шифрование уникального ключа промежуточным
	encrSK, err := c.encryptSK.GetActual()
	if err != nil {
		return model.Unit{}, err
	}
	unit.Meta.DataSK, err = encrypt(unitSK, encrSK.EncryptSK)
	if err != nil {
		return model.Unit{}, err
	}

	return unit, nil
}

func (c crypter) UnitDecrypt(unit model.Unit) (model.Unit, error) {
	// Дешифрование уникального ключа промежуточным
	encrSK, err := c.encryptSK.GetOld(unit.Meta.UploadedAt) // выбор из архива по дате
	if err != nil {
		return model.Unit{}, err
	}
	unitSK, err := decrypt(unit.Meta.DataSK, encrSK.EncryptSK)
	if err != nil {
		return model.Unit{}, err
	}
	// Дешифрование уникальным ключом
	decrString, err := decrypt(string(unit.Data), unitSK)
	if err != nil {
		return model.Unit{}, err
	}
	unit.Meta.DataSK = ""
	unit.Data = []byte(decrString)

	return unit, nil
}

func NewCrypter(cfg config.Config, store store.Store) (Crypter, error) {
	ctx := context.Background()

	// Получение промежуточных ключей из БД
	encrStrings, err := store.GetEncryptSK(ctx)
	if err != nil {
		return nil, err
	}

	// Дешифрование промежуточных ключей мастер-ключом
	var decrStrings []string
	for _, encrString := range encrStrings {
		decrString, err := decrypt(encrString, cfg.MasterSK)
		if err != nil {
			return nil, err
		}
		decrStrings = append(decrStrings, decrString) // JSON стркои
	}

	// Создание объекта промежуточных ключей
	encryptSK, err := NewEncryptSK(decrStrings)
	if err != nil {
		return nil, err
	}

	// Создание нового ключа
	key, err := encryptSK.GetActual()
	newKeyNeeded := false
	if err != nil {
		switch err {
		// Ключи отсутствуют
		case ErrNoEncryptKeys:
			newKeyNeeded = true
		default:
			return nil, err
		}
	} else {
		// Ключ устарел
		if (time.Since(key.Begin).Hours() / 24) > float64(cfg.NewSKIntervalD) {
			newKeyNeeded = true
		}
	}
	if newKeyNeeded {
		// Получение нового ключа в формате JSON
		newKey, err := encryptSK.CreateNewKey()
		if err != nil {
			return nil, err
		}
		// Шифрование мастер-ключом
		encrNewKey, err := encrypt(newKey, cfg.MasterSK)
		if err != nil {
			return nil, err
		}
		// Запись в БД
		err = store.SetEncryptSK(ctx, encrNewKey)
		if err != nil {
			return nil, err
		}
	}

	var crypter crypter
	crypter.cfg = cfg
	crypter.encryptSK = encryptSK
	return crypter, nil
}
