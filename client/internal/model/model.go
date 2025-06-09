package model

import (
	"time"
)

// Unit - модель единицы данных
type Unit struct {
	Name string   `json:"name"`
	Body UnitBody `json:"body"`
}

// UnitBody - тело единицы данных
type UnitBody struct {
	Meta UnitMeta `json:"meta"`
	Data []byte   `json:"data"`
}

// UnitMeta - метаданные единицы данных
type UnitMeta struct {
	Type       int       `json:"type"`
	ValidUntil time.Time `json:"validuntil"`
}

const (
	UnitTypeLogin  = 1
	UnitTypeText   = 2
	UnitTypeBinary = 3
	UnitTypeCard   = 4
)
