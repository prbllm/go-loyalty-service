package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/prbllm/go-loyalty-service/internal/config"
)

func NewZapLogger() (Logger, error) {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("02-01-2006 15:04:05.000")
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	zapConfig.DisableStacktrace = false
	zapConfig.DisableCaller = false

	logLevel, _ := os.LookupEnv(config.LogLevelEnv)
	switch strings.ToLower(logLevel) {
	case config.LogLevelDebug:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case config.LogLevelInfo:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case config.LogLevelWarn:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case config.LogLevelError:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case config.LogLevelFatal:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	zapLogger, err := zapConfig.Build(
		zap.AddStacktrace(zapcore.PanicLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return zapLogger.Sugar(), nil
}
