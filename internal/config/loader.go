package config

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Load загрузка конфига.
func Load() (*Config, error) {
	op := "config.loader"
	var (
		cfg     Config
		cfgPath string
		err     error
	)

	cfgPath = fetchConfigPath()

	err = cleanenv.ReadConfig(cfgPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &cfg, nil
}

// fetchConfigPath fetches config path from flag --config.
func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "../../config/local.yaml", "path to config file .yaml")
	flag.Parse()

	return path
}
