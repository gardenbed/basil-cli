package template

import "github.com/gardenbed/charm/ui"

// Service
type Service struct {
	ui ui.UI
}

func NewService(ui ui.UI) *Service {
	return &Service{
		ui: ui,
	}
}

// Execute executes all changes defined for a template.
func (s *Service) Execute(template Template) error {
	return template.Changes.execute(template.path, s.ui)
}
