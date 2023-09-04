package cmd

import (
	"github.com/sarulabs/di"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/gonah/cmd/migrate"
	diConfig "github.com/Kale-Grabovski/gonah/src/di"
	"github.com/Kale-Grabovski/gonah/src/domain"
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
	cobra.OnInitialize(initConfig, initDI, migrateDB)

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

func migrateDB() {
	conn := diContainer.Get("db").(domain.DB)
	logger := diContainer.Get("logger").(domain.Logger)

	migrator, err := migrate.NewMigrator(conn, "schema_version")
	if err != nil {
		logger.Error("Unable to create a migrator", zap.Error(err))
		return
	}

	err = migrator.LoadMigrations("./migrations")
	if err != nil {
		logger.Error("Unable to load migrations", zap.Error(err))
		return
	}

	err = migrator.Migrate(func(err error) (retry bool) {
		logger.Error("Commit failed during migration, retrying", zap.Error(err))
		return true
	})
	if err != nil {
		logger.Error("Unable to migrate", zap.Error(err))
		return
	}

	ver, err := migrator.GetCurrentVersion()
	if err != nil {
		logger.Error("Unable to get current schema version", zap.Error(err))
		return
	}

	logger.Info("Migration done. Current schema version", zap.Int32("ver", ver))
}
