package aesgcm

import (
	"testing"
	"time"

	"github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm/config"
	"github.com/iurnickita/gophkeeper1/internal/server/model"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
	"github.com/stretchr/testify/require"
)

func TestCrypter(t *testing.T) {

	tests := []struct {
		name string
		unit model.Unit
	}{
		{
			name: "test 1",
			unit: model.Unit{
				Key: model.UnitKey{
					UserID:   1,
					UnitName: "secret1",
				},
				Meta: model.UnitMeta{
					Type:       1,
					DataSK:     "",
					UploadedAt: time.Now().Add(time.Minute), // время должно быть позже создания ключа
				},
				Data: []byte("Таинственная тайна 1"),
			},
		},
	}

	// Config
	var cfg config.Config
	cfg.MasterSK = "cb459063d4bbbd4ce04a7c5b6e8121e7933630bada8fcb3abc20f6ca0aba3793"
	cfg.NewSKIntervalD = 30
	// Store
	store, _ := store.NewStoreMock()
	// Crypter
	crypter, err := NewCrypter(cfg, store)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Шифрование
			encrUnit, err := crypter.UnitEncrypt(test.unit)
			require.NoError(t, err)
			// Дешифрование
			decrUnit, err := crypter.UnitDecrypt(encrUnit)
			require.NoError(t, err)
			// Сравнение с исходным
			require.Equal(t, test.unit, decrUnit)
		})
	}

}
