package config

import (
	"os"
	"time"

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

type CacheConfig struct {
	TTLSeconds int `yaml:"ttlSeconds"`
}

type Config struct {
	Views []View      `yaml:"views"`
	Cache CacheConfig `yaml:"cache"`
}

// GetCacheTTL returns the cache TTL as a time.Duration
func (c *Config) GetCacheTTL() time.Duration {
	if c.Cache.TTLSeconds <= 0 {
		return 5 * time.Minute // Default to 5 minutes
	}
	return time.Duration(c.Cache.TTLSeconds) * time.Second
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
