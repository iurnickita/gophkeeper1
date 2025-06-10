// Пакет model. Модели данных
package model

import (
	"time"
)

// Unit - модель единицы данных
type Unit struct {
	Key  UnitKey
	Meta UnitMeta
	Data []byte
}

// UnitKey - ключ единицы данных
type UnitKey struct {
	UserID   int
	UnitName string
}

// UnitMeta - метаданные единицы данных
type UnitMeta struct {
	Type       int
	DataSK     string
	UploadedAt time.Time
}

const (
	UnitTypeLogin  = 1
	UnitTypeText   = 2
	UnitTypeBinary = 3
	UnitTypeCard   = 4
)
