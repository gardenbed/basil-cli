package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromFile(t *testing.T) {
	tests := []struct {
		name           string
		configFiles    []string
		expectedConfig Config
		expectedError  string
	}{
		{
			name:           "NoConfigFile",
			configFiles:    []string{"test/null"},
			expectedConfig: Config{},
		},
		{
			name:          "EmptyYAML",
			configFiles:   []string{"test/empty.yaml"},
			expectedError: "EOF",
		},
		{
			name:          "InvalidYAML",
			configFiles:   []string{"test/invalid.yaml"},
			expectedError: "cannot unmarshal",
		},
		{
			name:        "ValidYAML",
			configFiles: []string{"test/valid.yaml"},
			expectedConfig: Config{
				GPGKey: "0123456789ABCDEF",
				GitHub: GitHub{
					AccessToken: "ABCDEFGHIJKLMNOPQRSTabcdefghijklmnopqrst",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			configFiles = tc.configFiles
			config, err := FromFile()

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
