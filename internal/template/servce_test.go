package template

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/ui"
)

func TestNewService(t *testing.T) {
	ui := ui.NewNop()
	s := NewService(ui)

	assert.NotNil(t, s)
	assert.NotNil(t, s.ui)
}

func TestService_Execute(t *testing.T) {
	template := Template{
		path: "./test",
		Changes: Changes{
			Deletes:  Deletes{},
			Moves:    Moves{},
			Appends:  Appends{},
			Replaces: Replaces{},
		},
	}

	s := &Service{
		ui: ui.NewNop(),
	}

	err := s.Execute(template)

	assert.NoError(t, err)
}
