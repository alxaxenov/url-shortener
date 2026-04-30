// Package logger содержит настройку и глобальную сущность логгера.
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger = zap.NewNop().Sugar()

// Initialize инициализация логгера.
func Initialize() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("cannot initialize zap: %w", err)
	}
	defer logger.Sync()
	Logger = logger.Sugar()
	return nil
}
