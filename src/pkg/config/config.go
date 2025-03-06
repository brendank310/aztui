package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Action struct {
	Action      string `yaml:"action"`
	TakeFocus   bool   `yaml:"takeFocus"`
	Key         string `yaml:"key"`
	Width       int    `yaml:"width"`
	Description string `yaml:"description"`
}

type View struct {
	Name    string   `yaml:"view"`
	Actions []Action `yaml:"actions"`
}

type Config struct {
	Views []View `yaml:"views"`
}

var GConfig Config

func LoadConfig(configFile string) (Config, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	GConfig = config

	return config, err
}
