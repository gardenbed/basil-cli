package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupConfigFile(file string) (func(), error) {
	src, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(homeDir, configFiles[0])

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	cleanup := func() {
		_ = os.Remove(destPath)
	}

	return cleanup, nil
}

func TestRead(t *testing.T) {
	t.Run("NoConfigFile", func(t *testing.T) {
		configFiles = []string{"null"}
		config, err := Read()
		assert.NoError(t, err)
		assert.Equal(t, Config{}, config)
	})

	tests := []struct {
		name           string
		configFile     string
		expectedConfig Config
		expectedError  string
	}{
		{
			name:           "EmptyConfigFile",
			configFile:     "test/empty.yaml",
			expectedConfig: Config{},
			expectedError:  "EOF",
		},
		{
			name:           "InvalidConfigFile",
			configFile:     "test/invalid.yaml",
			expectedConfig: Config{},
			expectedError:  "yaml: unmarshal errors",
		},
		{
			name:       "Success",
			configFile: "test/valid.yaml",
			expectedConfig: Config{
				GitHub: GitHub{
					AccessToken: "ABCDEFGHIJKLMNOPQRSTabcdefghijklmnopqrst",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configFiles = []string{".basil.test.yaml"}
			cleanup, err := setupConfigFile(tc.configFile)
			assert.NoError(t, err)
			defer cleanup()

			config, err := Read()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Equal(t, Config{}, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedConfig, config)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Run("EmptyConfig", func(t *testing.T) {
		path, err := Write(Config{})

		assert.Empty(t, path)
		assert.NoError(t, err)
	})

	t.Run("InvalidFile", func(t *testing.T) {
		configFiles = []string{"."}
		path, err := Write(Config{
			GitHub: GitHub{
				AccessToken: "access_token",
			},
		})

		assert.Empty(t, path)
		assert.EqualError(t, err, "open /Users/milad: is a directory")
	})

	t.Run("Success", func(t *testing.T) {
		configFiles = []string{".basil.test.yaml"}
		path, err := Write(Config{
			GitHub: GitHub{
				AccessToken: "access_token",
			},
		})

		defer os.Remove(path)

		assert.NotEmpty(t, path)
		assert.NoError(t, err)
	})
}
