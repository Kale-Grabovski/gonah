package domain

type Config struct {
	LogLevel string `yaml:"logLevel"`
	Listen   string `yaml:"listen"`
	DB       struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
}
