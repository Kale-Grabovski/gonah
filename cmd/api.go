package cmd

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Kale-Grabovski/gonah/src/api"
	"github.com/Kale-Grabovski/gonah/src/domain"
)

var apiCmd = &cobra.Command{
	Use: "api",
	Run: func(cmd *cobra.Command, args []string) {
		runApi()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func runApi() {
	e := echo.New()

	users := diContainer.Get("api.users").(*api.UsersAction)
	e.GET("/users", users.GetAll)
	e.GET("/users/:id", users.GetById)
	e.POST("/users", users.Create)
	e.DELETE("/users/:id", users.Delete)

	logger := diContainer.Get("logger").(domain.Logger)
	logger.Info("Starting API")

	err := e.Start(viper.GetString("listen"))
	if err != nil {
		panic(err)
	}
}
