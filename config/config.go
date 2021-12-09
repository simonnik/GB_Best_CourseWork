package config

import (
	"flag"

	"github.com/kkyr/fig"
)

// Config - структура для конфигурации
type Config struct {
	App struct {
		Timeout int    `fig:"timeout" default:"5"`
		File    string `fig:"config" default:"config.yaml"`
	} `yaml:"app"`
}

func NewConfig() (*Config, error) {
	var (
		cfg Config
	)
	configFile := flag.String("config", "config.yaml", "set path to configuration file")
	flag.Parse()
	err := fig.Load(&cfg, fig.File(*configFile))
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
