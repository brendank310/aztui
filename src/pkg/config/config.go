package config

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

type Action struct {
	Type      string `yaml:"type"`
	Condition string `yaml:"condition"`
	Action    string `yaml:"action"`
	TakeFocus bool   `yaml:"takeFocus"`
}

type KeyMapping struct {
	Action string `yaml:"action"`
	Key    string `yaml:"key"`
}

type Config struct {
	Actions     []Action     `yaml:"actions"`
	KeyMappings []KeyMapping `yaml:"key_mappings"`
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

func MapUserKeyToEvent(userKey string) (tcell.Key, tcell.ModMask) {
	keyParts := strings.Split(userKey, "-")
	var key tcell.Key
	var mod tcell.ModMask

	for _, part := range keyParts {
		switch part {
		case "Ctrl":
			mod |= tcell.ModCtrl
		case "Alt":
			mod |= tcell.ModAlt
		case "R":
			key = tcell.KeyCtrlR
		case "F5":
			key = tcell.KeyF5
		case "C":
			key = tcell.KeyCtrlC
		}
	}

	return key, mod
}
