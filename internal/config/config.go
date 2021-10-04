package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var (
	configFiles = []string{"~/.basil.yml", "~/.basil.yaml"}
)

// Config is the model for all configurations.
type Config struct {
	GPGKey string `yaml:"gpg_key"`
	GitHub GitHub `yaml:"github"`
}

// GitHub has the configurations for GitHub.
type GitHub struct {
	AccessToken string `yaml:"access_token"`
}

// FromFile reads the Basil configuration file in user's home directory.
// If no config file is found, an empty config will be returned.
func FromFile() (Config, error) {
	var config Config

	for _, configFile := range configFiles {
		file, err := os.Open(configFile)
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
