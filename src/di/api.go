package di

import (
	"github.com/sarulabs/di"

	"github.com/Kale-Grabovski/gonah/src/api"
	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/repo"
)

var ConfigApi = []di.Def{
	{
		Name:  "api.users",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			usersRepo := ctx.Get("repo.user").(*repo.UserRepo)
			logger := ctx.Get("logger").(domain.Logger)
			return api.NewUsersAction(usersRepo, logger), nil
		},
	},
}
