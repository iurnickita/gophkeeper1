package aesgcm

import (
	"testing"
	"time"

	"github.com/iurnickita/gophkeeper1/internal/server/model"
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
					UploadedAt: time.Now(),
				},
				Data: []byte("Таинственная тайна 1"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			/* // Шифрование
			encrUnit, err := UnitEncrypt(test.unit)
			require.NoError(t, err)
			// Дешифрование
			decrUnit, err := UnitDecrypt(encrUnit)
			require.NoError(t, err)
			// Сравнение с исходным
			require.Equal(t, test.unit, decrUnit) */
		})
	}

}
