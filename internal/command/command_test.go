package command

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunPreflightChecks(t *testing.T) {
	tests := []struct {
		name          string
		environment   map[string]string
		ctx           context.Context
		checklist     PreflightChecklist
		expectedError string
	}{
		{
			name:        "NoCheck",
			environment: map[string]string{},
			ctx:         context.Background(),
			checklist:   PreflightChecklist{},
		},
		{
			name:        "AllChecks",
			environment: map[string]string{},
			ctx:         context.Background(),
			checklist: PreflightChecklist{
				Go: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.environment {
				assert.NoError(t, os.Setenv(k, v))
				defer os.Unsetenv(k)
			}

			info, err := RunPreflightChecks(tc.ctx, tc.checklist)

			if tc.expectedError != "" {
				assert.Zero(t, info)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, info)
				assert.NotEmpty(t, info.WorkingDirectory)
			}
		})
	}
}
