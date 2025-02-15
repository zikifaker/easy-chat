package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	ErrUnableToReadConfigFile  = errors.New("unable to read config file")
	ErrUnableToParseConfigFile = errors.New("unable to parse config file")
)

var globalConfig *Config

type DataBaseConfig struct {
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	DataBase struct {
		Mysql DataBaseConfig `yaml:"mysql"`
	} `yaml:"database"`
	SecretKey struct {
		JWT string `yaml:"jwt"`
	} `yaml:"secret_key"`
	APIKey struct {
		Qwen string `yaml:"qwen"`
		Exa  string `yaml:"exa"`
	} `yaml:"api_key"`
}

func Init(configPath string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnableToReadConfigFile, err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(content, config); err != nil {
		return fmt.Errorf("%w: %v", ErrUnableToParseConfigFile, err)
	}

	globalConfig = config
	return nil
}

func Get() *Config {
	return globalConfig
}
