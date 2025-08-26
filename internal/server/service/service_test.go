package service

import (
	"context"
	"testing"
	"time"

	"github.com/iurnickita/gophkeeper1/internal/server/config"
	"github.com/iurnickita/gophkeeper1/internal/server/crypto/aesgcm"
	"github.com/iurnickita/gophkeeper1/internal/server/logger"
	"github.com/iurnickita/gophkeeper1/internal/server/model"
	"github.com/iurnickita/gophkeeper1/internal/server/store"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
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
	cfg.Crypter.MasterSK = "cb459063d4bbbd4ce04a7c5b6e8121e7933630bada8fcb3abc20f6ca0aba3793"
	cfg.Crypter.NewSKIntervalD = 30
	cfg.Logger.LogLevel = "debug"
	// Store
	store, _ := store.NewStoreMock()
	// Crypter
	crypter, err := aesgcm.NewCrypter(cfg.Crypter, store)
	require.NoError(t, err)
	// Logger
	logger, err := logger.NewZapLog(cfg.Logger)
	require.NoError(t, err)
	// Service
	service, err := NewService(cfg.Service, store, crypter, logger)
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			// Write
			err = service.Write(ctx, test.unit)
			require.NoError(t, err)
			// Read
			respUnit, err := service.Read(
				ctx, test.unit.Key.UserID, test.unit.Key.UnitName)
			require.NoError(t, err)
			require.Equal(t, test.unit, respUnit)
		})
	}
}
