package di

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sarulabs/di"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

var ConfigCommon = []di.Def{
	{
		Name:  "db",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			cfg := ctx.Get("config").(*domain.Config)
			conn, err := pgxpool.Connect(context.Background(), cfg.DB.DSN)
			if err != nil {
				panic(err)
			}
			return conn, err
		},
		Close: func(obj interface{}) error {
			obj.(*pgxpool.Pool).Close()
			return nil
		},
	},
	{
		Name:  "logger",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			var conf = zap.NewProductionConfig()
			cfg := ctx.Get("config").(*domain.Config)
			err := conf.Level.UnmarshalText([]byte(cfg.LogLevel))
			if err != nil {
				return nil, err
			}
			conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			return conf.Build()
		},
		Close: func(obj interface{}) error {
			return obj.(*zap.Logger).Sync()
		},
	},
}
