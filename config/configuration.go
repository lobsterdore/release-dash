package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Github github
	Server server
}

type github struct {
	Pat string `env:"GITHUB_PAT" envDefault:""`
}

type server struct {
	Host    string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port    string `env:"SERVER_PORT" envDefault:"8080"`
	Timeout serverTimeout
}

type serverTimeout struct {
	Server int `env:"SERVER_TIMEOUT_SERVER" envDefault:"180"`
	Write  int `env:"SERVER_TIMEOUT_READ" envDefault:"180"`
	Read   int `env:"SERVER_TIMEOUT_WRITE" envDefault:"180"`
	Idle   int `env:"SERVER_TIMEOUT_IDLE" envDefault:"180"`
}

func NewConfig() (Config, error) {

	cfg := &Config{}
	err := env.Parse(cfg)

	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}
