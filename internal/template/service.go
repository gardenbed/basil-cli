package template

import "github.com/gardenbed/basil-cli/internal/debug"

// Service
type Service struct {
	debugger *debug.DebuggerSet
}

func NewService(level debug.Level) *Service {
	return &Service{
		debugger: debug.NewSet(level),
	}
}

// Execute executes all changes defined for a template.
func (s *Service) Execute(template Template) error {
	return template.Changes.execute(template.path, s.debugger)
}
