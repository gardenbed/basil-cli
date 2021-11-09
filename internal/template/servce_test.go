package template

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gardenbed/basil-cli/internal/debug"
)

func TestNewService(t *testing.T) {
	s := NewService(debug.None)

	assert.NotNil(t, s)
	assert.NotNil(t, s.debugger)
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
		debugger: debug.NewSet(debug.None),
	}

	err := s.Execute(template)

	assert.NoError(t, err)
}
