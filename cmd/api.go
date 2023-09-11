package cmd

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Kale-Grabovski/gonah/src/api"
	"github.com/Kale-Grabovski/gonah/src/domain"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

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
	e.Validator = &CustomValidator{validator: validator.New()}

	users := diContainer.Get("api.users").(*api.UsersAction)
	e.GET("/api/v1/users", users.GetAll)
	e.GET("/api/v1/users/:id", users.GetById)
	e.POST("/api/v1/users", users.Create)
	e.DELETE("/api/v1/users/:id", users.Delete)

	logger := diContainer.Get("logger").(domain.Logger)
	logger.Info("Starting API")

	go func() {
		err := e.Start(":" + viper.GetString("listen"))
		if err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
