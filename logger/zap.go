package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var levelMap = map[string]zapcore.Level{
	"debug":  zap.DebugLevel,
	"info":   zap.InfoLevel,
	"warn":   zap.WarnLevel,
	"error":  zap.ErrorLevel,
	"dpanic": zap.DPanicLevel,
	"panic":  zap.PanicLevel,
	"fatal":  zap.FatalLevel,
}

func NewLogger(config *Config, opts ...zap.Option) *zap.Logger {
	if config == nil {
		panic("Initialize logger error: config is nil")
	}

	if config.Level == "" {
		config.Level = "info"
	}

	if config.Name == "" {
		config.Name = os.Getenv("HOSTNAME")
	}

	if config.Encoding == "" {
		config.Encoding = "console"
	}

	if config.Name != "" {
		opts = append(opts, zap.Fields(zap.String("machine", config.Name)))
	}

	level, ok := levelMap[config.Level]
	if !ok {
		panic(fmt.Sprintf("Initialize logger error: wrong logger level [%s]", config.Level))
	}

	encConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		MessageKey:     "_msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         config.Encoding,
		EncoderConfig:    encConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	ret, err := zapConfig.Build(opts...)
	if err != nil {
		panic(fmt.Sprintf("Initialize logger error: %v", err))
	}
	zap.ReplaceGlobals(ret)
	return ret
}
