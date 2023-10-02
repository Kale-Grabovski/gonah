package di

import (
	"github.com/sarulabs/di"

	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/service"
)

var ConfigService = []di.Def{
	{
		Name:  "service.kafka",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			cfg := ctx.Get("config").(*domain.Config)
			logger := ctx.Get("logger").(domain.Logger)
			return service.NewKafka(cfg, logger), nil
		},
		Close: func(obj interface{}) error {
			obj.(*service.Kafka).Close()
			return nil
		},
	},
}
