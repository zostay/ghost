package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = ".ghost.yaml"

var cfgInstance *Config

// KeeperConfig is an unstructured map of configuration values for a keeper.
type KeeperConfig map[string]any

// Config is the top-level configuration for ghost. When .ghost.yaml is loaded,
// this is the structure it must match.
type Config struct {
	MasterKeeper string                  `yaml:"master"`
	Keepers      map[string]KeeperConfig `yaml:"keepers"`
}

// configPath locates teh configuration file.
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

// New creates a new, empty configuration.
func New() *Config {
	return &Config{
		Keepers: map[string]KeeperConfig{},
	}
}

// Instance returns the singleton instance of the configuration.
func Instance() *Config {
	if cfgInstance != nil {
		return cfgInstance
	}

	cfgInstance = New()
	return cfgInstance
}

// Load loads the configuration from the given path. If the path is empty, the
// default path is used.
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

// Save saves the configuration to the given path. If the path is empty, the
// default path is used.
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
