package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = ".ghost.yaml"

type Config struct {
	Keepers map[string]KeeperConfig `yaml:"keepers"`
}

type KeeperConfig struct {
	LastPass LastPassConfig    `yaml:"lastpass"`
	Keepass  KeepassConfig     `yaml:"keepass"`
	Low      LowSecurityConfig `yaml:"low"`
	Router   RouterConfig      `yaml:"router"`
	Seq      SeqConfig         `yaml:"seq"`
}

type LastPassConfig struct {
	Username string `yaml:"username"`
}

type KeepassConfig struct {
	Path string `yaml:"path"`
}

type LowSecurityConfig struct {
	Path string `yaml:"path"`
}

type RouterConfig struct {
	Routes       []RouteConfig `yaml:"routes"`
	DefaultRoute string        `yaml:"default"`
}

type RouteConfig struct {
	Locations []string `yaml:"locations"`
	Keeper    string   `yaml:"keeper"`
}

type SeqConfig struct {
	Keepers []string `yaml:"keepers"`
}

func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, configFile), nil
}

func LoadConfig() (*Config, error) {
	cp, err := configPath()
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(cp)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	} else if stat.Mode().Perm()&0o77 != 0 {
		return nil, fmt.Errorf("config file %q has insecure permissions", cp)
	}

	configData, err := os.ReadFile(cp)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(c *Config) error {
	cp, err := configPath()
	if err != nil {
		return err
	}

	configData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(cp, configData, 0600)
	if err != nil {
		return err
	}

	return nil
}
