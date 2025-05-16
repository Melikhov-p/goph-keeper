package config

import (
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	masterKeyByteLen = 32
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

	err = cfg.GetMasterKey()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get master key with error %w", op, err)
	}

	return &cfg, nil
}

// fetchConfigPath fetches config path from flag --config.
func fetchConfigPath() string {
	var path string

	flag.StringVar(
		&path,
		"config",
		"C:\\Users\\melik\\GolandProjects\\goph-keeper\\config\\local.yaml",
		"path to config file .yaml",
	)
	flag.Parse()

	return path
}

// GetMasterKey получение мастер-ключа для (де-)шифрования.
func (c *Config) GetMasterKey() error {
	op := "config.Loader.GetMasterKey"

	var err error

	c.Security.MasterKey, err = hex.DecodeString("f8f2761b99775dac26e373e4942d6fd648f29325db7312158cc88205ff5e86b8")
	if err != nil {
		return fmt.Errorf("%s: failed to decode string for master key with error %w", op, err)
	}

	if len(c.Security.MasterKey) != masterKeyByteLen {
		return fmt.Errorf("decoded master key has invalid length: %d bytes, expected 32", len(c.Security.MasterKey))
	}

	return nil
}
