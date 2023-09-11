package domain

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Panic(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Debug(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
}

func NewLogger() (*zap.Logger, error) {
	var conf = zap.NewProductionConfig()
	_ = conf.Level.UnmarshalText([]byte(viper.GetString("loglevel")))
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return conf.Build()
}
