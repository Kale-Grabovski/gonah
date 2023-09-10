package di

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sarulabs/di"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/src/domain"
)

var ConfigCommon = []di.Def{
	{
		Name:  "db",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			conn, err := pgxpool.Connect(context.Background(), viper.GetString("db.dsn"))
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
			return domain.NewLogger()
		},
		Close: func(obj interface{}) error {
			return obj.(*zap.Logger).Sync()
		},
	},
}
