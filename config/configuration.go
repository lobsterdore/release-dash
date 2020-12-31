package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
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
	Level string `env:"LOGGING_LEVEL" envDefault:"info"`
}

type server struct {
	Host    string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port    string `env:"SERVER_PORT" envDefault:"8080"`
	Timeout serverTimeout
}

type serverTimeout struct {
	Idle   int `env:"SERVER_TIMEOUT_IDLE" envDefault:"65"`
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
