package domain

import "go.uber.org/zap/zapcore"

type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Debug(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
}
