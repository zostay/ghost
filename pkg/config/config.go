package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = ".ghost.yaml"

var cfgInstance *Config

type KeeperConfig map[string]any

type Config struct {
	MasterKeeper string                  `yaml:"master"`
	Keepers      map[string]KeeperConfig `yaml:"keepers"`
}

func configPath(requestedPath string) (string, error) {
	if requestedPath != "" {
		requestedDir := filepath.Dir(requestedPath)
		if info, err := os.Stat(requestedDir); os.IsNotExist(err) || !info.IsDir() {
			return "", fmt.Errorf("requested configuration path directory %q does not exist", requestedDir)
		}

		return requestedPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, configFile), nil
}

func New() *Config {
	return &Config{
		Keepers: map[string]KeeperConfig{},
	}
}

func Instance() *Config {
	if cfgInstance != nil {
		return cfgInstance
	}

	cfgInstance = New()
	return cfgInstance
}

func (c *Config) Load(requestedPath string) error {
	cp, err := configPath(requestedPath)
	if err != nil {
		return err
	}

	_, err = os.Stat(cp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	configData, err := os.ReadFile(cp)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configData, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Save(requestedPath string) error {
	cp, err := configPath(requestedPath)
	if err != nil {
		return err
	}

	cf, err := os.Create(cp)
	if err != nil {
		return err
	}
	defer cf.Close()

	enc := yaml.NewEncoder(cf)
	defer enc.Close()
	enc.SetIndent(2)
	err = enc.Encode(c)
	if err != nil {
		return err
	}

	return nil
}
