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
	AccessToken string `yaml:"access_token" ask:"secret, your personal access token"`
}

func findFile(useDefault bool) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	for _, configFile := range configFiles {
		path := filepath.Join(homeDir, configFile)
		_, err := os.Stat(path)

		if err == nil {
			return path, nil
		}

		if !os.IsNotExist(err) {
			return "", err
		}
	}

	if useDefault {
		return filepath.Join(homeDir, configFiles[0]), nil
	}

	return "", nil
}

// Read reads the Basil configurations from a file in user's home directory.
// If no config file is found, an empty config will be returned.
func Read() (Config, error) {
	path, err := findFile(false)
	if err != nil {
		return Config{}, err
	}

	// If no config file found, return an empty config object
	if path == "" {
		return Config{}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Write writes the Basil configurations into a file in user's home directory.
// If the config is empty, no config file will be written.
func Write(config Config) (string, error) {
	if config == (Config{}) {
		return "", nil
	}

	path, err := findFile(true)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}

	if err := yaml.NewEncoder(file).Encode(config); err != nil {
		return "", err
	}

	return path, nil
}
