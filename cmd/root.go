package cmd

import (
	"fmt"
	"strings"

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
	cobra.OnInitialize(initConfigAndDI, func() {
		conn := diContainer.Get("db").(domain.DB)
		logger := diContainer.Get("logger").(domain.Logger)
		migrate.Run("./migrations", conn, logger)
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfigAndDI() {
	builder, _ := di.NewBuilder()
	cfgDI := di.Def{
		Name:  "config",
		Scope: di.App,
		Build: func(ctx di.Container) (interface{}, error) {
			return initConfig()
		},
	}
	err := builder.Add(append([]di.Def{cfgDI}, diConfig.Config...)...)
	if err != nil {
		panic("Unable to build DI containers: " + err.Error())
	}
	diContainer = builder.Build()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() (*domain.Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetEnvPrefix(domain.EnvPrefix)

	if cfgFile == "" {
		cfgFile = "config-example.yaml"
	}
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error occurred while reading config file: %v", err)
	}
	var config *domain.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config file: %v", err)
	}
	return config, nil
}
