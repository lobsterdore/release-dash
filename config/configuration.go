package config

import (
	"github.com/caarlos0/env/v6"
)

type cache struct {
	CleanupIntervalSeconds   int `env:"CACHE_CLEANUP_INTERVAL_SECONDS" envDefault:"60"`
	DefaultExpirationSeconds int `env:"CACHE_DEFAULT_EXPIRATION_SECONDS" envDefault:"1800"`
}

type Config struct {
	Cache   cache
	Github  github
	Logging logging
	Server  server
}

type github struct {
	ChangelogFetchTimerSeconds int    `env:"GITHUB_CHANGELOG_FETCH_TIMER_SECONDS" envDefault:"60"`
	Pat                        string `env:"GITHUB_PAT" envDefault:""`
	RepoFetchTimerSeconds      int    `env:"GITHUB_REPO_FETCH_TIMER_SECONDS" envDefault:"600"`
	UrlDefault                 string `env:"GITHUB_URL_DEFAULT" envDefault:""`
	UrlUpload                  string `env:"GITHUB_URL_UPLOAD" envDefault:""`
}

type logging struct {
	Level string `env:"LOGGING_LEVEL" envDefault:"error"`
}

type server struct {
	Host    string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port    string `env:"SERVER_PORT" envDefault:"8080"`
	Timeout serverTimeout
}

type serverTimeout struct {
	Idle   int `env:"SERVER_TIMEOUT_IDLE" envDefault:"65"`
	Read   int `env:"SERVER_TIMEOUT_WRITE" envDefault:"10"`
	Server int `env:"SERVER_TIMEOUT_SERVER" envDefault:"10"`
	Write  int `env:"SERVER_TIMEOUT_READ" envDefault:"10"`
}

func NewConfig() (Config, error) {

	cfg := &Config{}
	err := env.Parse(cfg)

	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}
