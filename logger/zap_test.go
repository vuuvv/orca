package logger

import (
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
	"os"
	"testing"
)

func TestZap(t *testing.T) {
	_ = os.Setenv("HOSTNAME", "test-machine")
	log := NewLogger(&Config{Encoding: "json"})
	log.Info("hello", zap.Stack("key"))
	log.Info("error", zap.Error(errors.New("error")))
	log.Error("error", zap.Error(errors.New("error")))

	logger := NewLogger(&Config{Encoding: "console"})
	logger.Error("error", zap.Error(errors.New("error")))
}
