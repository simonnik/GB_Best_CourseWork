package config

import (
	"github.com/kkyr/fig"
)

// Config - структура для конфигурации
type Config struct {
	App struct {
		Timeout int    `fig:"timeout" default:"5" yaml:"timeout"`
		File    string `fig:"config" default:"config.yaml"`
	} `yaml:"app"`
}

func NewConfig(configFile *string) (*Config, error) {
	var (
		cfg Config
	)
	err := fig.Load(&cfg, fig.File(*configFile))
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
