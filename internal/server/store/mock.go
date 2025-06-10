package store

import (
	"context"
	"errors"

	"github.com/iurnickita/gophkeeper1/internal/server/model"
)

type storeMock struct {
	encryptionSK []string
	dataUnits    map[string]model.Unit
}

// AuthLogin implements Store.
func (s *storeMock) AuthLogin(ctx context.Context, login string, password string) (int, error) {
	panic("unimplemented")
}

// AuthRegister implements Store.
func (s *storeMock) AuthRegister(ctx context.Context, login string, password string) (int, error) {
	panic("unimplemented")
}

// List implements Store.
func (s *storeMock) List(ctx context.Context, userID int) ([]string, error) {
	panic("unimplemented")
}

// Read implements Store.
func (s *storeMock) Read(ctx context.Context, userID int, unitName string) (model.Unit, error) {
	unit, ok := s.dataUnits[unitName]
	if !ok {
		return model.Unit{}, errors.New("mock: no rows")
	}
	return unit, nil
}

// Write implements Store.
func (s *storeMock) Write(ctx context.Context, unit model.Unit) error {
	s.dataUnits[unit.Key.UnitName] = unit
	return nil
}

// Delete implements Store.
func (s *storeMock) Delete(ctx context.Context, userID int, unitName string) error {
	panic("unimplemented")
}

// GetEncryptSK implements Store.
func (s *storeMock) GetEncryptSK(ctx context.Context) ([]string, error) {
	return s.encryptionSK, nil
}

// SetEncryptSK implements Store.
func (s *storeMock) SetEncryptSK(ctx context.Context, sk string) error {
	s.encryptionSK = append(s.encryptionSK, sk)
	return nil
}

// NewStoreMock создает mock хранилища
func NewStoreMock() (Store, error) {
	dataUnits := make(map[string]model.Unit)
	return &storeMock{dataUnits: dataUnits}, nil
}
