package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	configFiles = []string{".basil.yml", ".basil.yaml"}
)

// Config is the model for all configurations.
type Config struct {
	GitHub GitHub `yaml:"github"`
}

// GitHub has the configurations for GitHub.
type GitHub struct {
	AccessToken string `yaml:"access_token"`
}

// Read reads the Basil configurations from a file in user's home directory.
// If no config file is found, an empty config will be returned.
func Read() (Config, error) {
	var config Config

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	for _, configFile := range configFiles {
		file, err := os.Open(filepath.Join(homeDir, configFile))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Config{}, err
		}
		defer file.Close()

		if err := yaml.NewDecoder(file).Decode(&config); err != nil {
			return Config{}, err
		}

		break
	}

	return config, nil
}

// Write writes the Basil configurations into a file in user's home directory.
// If the config is empty, no config file will be written.
func Write(config Config) (string, error) {
	if config == (Config{}) {
		return "", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(homeDir, configFiles[0])
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", err
	}

	if err := yaml.NewEncoder(file).Encode(config); err != nil {
		return "", err
	}

	return path, nil
}
