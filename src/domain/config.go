package domain

const EnvPrefix = "GONAH"

type Config struct {
	LogLevel string `yaml:"logLevel"`
	ApiPort  string `yaml:"apiPort"`
	DB       struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
	Kafka struct {
		Host string `yaml:"host"`
	} `yaml:"kafka"`
}
