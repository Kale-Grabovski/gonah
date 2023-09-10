package cmd

import (
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	diConfig "github.com/Kale-Grabovski/gonah/src/di"
	"github.com/Kale-Grabovski/gonah/src/domain"
	"github.com/Kale-Grabovski/gonah/src/service/migrate"
)

var cfgFile string
var diContainer di.Container

var rootCmd = &cobra.Command{
	Use: "gonah",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initDI, func() {
		conn := diContainer.Get("db").(domain.DB)
		logger := diContainer.Get("logger").(domain.Logger)
		migrate.Run("./migrations", conn, logger)
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		cfgFile = "config.yaml"
	}

	viper.AutomaticEnv()
	viper.SetConfigFile(cfgFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic("Error occurred while reading config file\n")
	}
}

func initDI() {
	builder, _ := di.NewBuilder()
	err := builder.Add(diConfig.Config...)
	if err != nil {
		panic("Unable to build DI containers: " + err.Error())
	}
	diContainer = builder.Build()
}
