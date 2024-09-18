package config

import (
	"fmt"
	"os"

	"github.com/brendank310/aztui/pkg/logger"
	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v3"
)

type UserKey struct {
	Key tcell.Key
	Ch  rune
}

func (k *UserKey) UnmarshalYAML(value *yaml.Node) error {
	var key string
	if err := value.Decode(&key); err != nil {
		return err
	}

	tcellKey, runeValue, err := MapUserKeyToEvent(key)
	if err != nil {
		return err
	}

	k.Key = tcellKey
	k.Ch = runeValue

	return nil
}

type Action struct {
	Action    string  `yaml:"action"`
	TakeFocus bool    `yaml:"takeFocus"`
	Key       UserKey `yaml:"key"`
	Width     int     `yaml:"width"`
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

func MapUserKeyToEvent(userKey string) (tcell.Key, rune, error) {
	NameKeys := make(map[string]tcell.Key)
	for key, name := range tcell.KeyNames {
		NameKeys[name] = key
	}

	// check if the key is a named key
	if key, exists := NameKeys[userKey]; exists {
		logger.Println("Mapped key", userKey, "to", key, "0")
		return key, 0, nil
	}

	if len(userKey) == 1 {
		logger.Println("Mapped runekey", userKey, "to", tcell.KeyRune)
		return tcell.KeyRune, rune(userKey[0]), nil
	}

	return 0, 0, fmt.Errorf("unable to map key %s to a tcell key", userKey)
}
