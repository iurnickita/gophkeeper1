// Пакет logger. Журнал
package logger

import (
	"go.uber.org/zap"

	"github.com/iurnickita/gophkeeper1/internal/client/logger/config"
)

// NewZapLog создает объект zap-логгера
func NewZapLog(cfg config.Config) (*zap.Logger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логгера
	zapcfg := zap.NewProductionConfig()
	// устанавливаем уровень
	zapcfg.Level = lvl
	// создаём логгер на основе конфигурации
	zl, err := zapcfg.Build()
	if err != nil {
		return nil, err
	}
	//
	return zl, nil
}
