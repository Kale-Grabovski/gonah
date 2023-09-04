package di

import (
	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/repo"

	"github.com/sarulabs/di"
)

var ConfigRepo = []di.Def{
	{
		Name:  "repo.user",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			db := ctx.Get("db").(domain.DB)
			return repo.NewUserRepository(db), nil
		},
	},
}
