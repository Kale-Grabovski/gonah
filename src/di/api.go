package di

import (
	"github.com/sarulabs/di"

	"github.com/Kale-Grabovski/gonah/src/api"
	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/repo"
	"github.com/Kale-Grabovski/gonah/src/service"
)

var ConfigApi = []di.Def{
	{
		Name:  "api.users",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			usersRepo := ctx.Get("repo.user").(*repo.UserRepo)
			logger := ctx.Get("logger").(domain.Logger)
			kaf := ctx.Get("service.kafka").(*service.Kafka)
			return api.NewUsersAction(usersRepo, kaf, logger), nil
		},
	},
}
