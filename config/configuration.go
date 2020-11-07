package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Cache  cache
	Github github
	Server server
}

type cache struct {
	CleanupIntervalMinutes   int `env:"CACHE_CLEANUP_INTERVAL_MINUTES" envDefault:"5"`
	DefaultExpirationMinutes int `env:"CACHE_DEFAULT_EXPIRATION_MINUTES" envDefault:"10"`
}

type github struct {
	FetchTimerSeconds int    `env:"GITHUB_FETCH_TIMER_SECONDS" envDefault:"900"`
	Pat               string `env:"GITHUB_PAT" envDefault:""`
}

type server struct {
	Host    string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port    string `env:"SERVER_PORT" envDefault:"8080"`
	Timeout serverTimeout
}

type serverTimeout struct {
	Idle   int `env:"SERVER_TIMEOUT_IDLE" envDefault:"90"`
	Read   int `env:"SERVER_TIMEOUT_WRITE" envDefault:"30"`
	Server int `env:"SERVER_TIMEOUT_SERVER" envDefault:"30"`
	Write  int `env:"SERVER_TIMEOUT_READ" envDefault:"30"`
}

func NewConfig() (Config, error) {

	cfg := &Config{}
	err := env.Parse(cfg)

	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}
