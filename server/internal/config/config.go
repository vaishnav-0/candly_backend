package config

import (
	"candly/internal/db"
	"candly/internal/logging"
	"candly/internal/memstore"
	"github.com/caarlos0/env/v7"
)

var base_path string = "./"
var file_name string

type Mode string

const (
	Development Mode = "development"
	Production  Mode = "production"
)

type Config struct {
	Mode    Mode `env:"MODE,notEmpty"`
	SwaggerAPIKey string `env:"SWAGGER_API_KEY,notEmpty"`
	Db      db.Config
	Redis   memstore.Config
	Logging logging.Config
	JWTKey string `env:"JWT_KEY"`
	JWTPub string `env:"JWT_PUB"`
}

func GetConfig() Config {

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}
