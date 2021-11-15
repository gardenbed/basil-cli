package config

import "github.com/gardenbed/basil-cli/internal/ui"

type (
	ConfirmMock struct {
		InPrompt     string
		InDefault    bool
		OutConfirmed bool
		OutError     error
	}

	AskMock struct {
		InPrompt   string
		InDefault  string
		InValidate ui.ValidateFunc
		OutValue   string
		OutError   error
	}

	AskSecretMock struct {
		InPrompt   string
		InConfirm  bool
		InValidate ui.ValidateFunc
		OutValue   string
		OutError   error
	}

	SelectMock struct {
		InPrompt string
		InSize   int
		InItems  []ui.Item
		InSearch ui.SearchFunc
		OutItem  ui.Item
		OutError error
	}

	MockUI struct {
		ui.UI

		ConfirmIndex int
		ConfirmMocks []ConfirmMock

		AskIndex int
		AskMocks []AskMock

		AskSecretIndex int
		AskSecretMocks []AskSecretMock

		SelectIndex int
		SelectMocks []SelectMock
	}
)

func (m *MockUI) Confrim(prompt string, Default bool) (bool, error) {
	i := m.ConfirmIndex
	m.ConfirmIndex++
	m.ConfirmMocks[i].InPrompt = prompt
	m.ConfirmMocks[i].InDefault = Default
	return m.ConfirmMocks[i].OutConfirmed, m.ConfirmMocks[i].OutError
}

func (m *MockUI) Ask(prompt, Default string, validate ui.ValidateFunc) (string, error) {
	i := m.AskIndex
	m.AskIndex++
	m.AskMocks[i].InPrompt = prompt
	m.AskMocks[i].InDefault = Default
	m.AskMocks[i].InValidate = validate
	return m.AskMocks[i].OutValue, m.AskMocks[i].OutError
}

func (m *MockUI) AskSecret(prompt string, confirm bool, validate ui.ValidateFunc) (string, error) {
	i := m.AskSecretIndex
	m.AskSecretIndex++
	m.AskSecretMocks[i].InPrompt = prompt
	m.AskSecretMocks[i].InConfirm = confirm
	m.AskSecretMocks[i].InValidate = validate
	return m.AskSecretMocks[i].OutValue, m.AskSecretMocks[i].OutError
}

func (m *MockUI) Select(prompt string, size int, items []ui.Item, search ui.SearchFunc) (ui.Item, error) {
	i := m.SelectIndex
	m.SelectIndex++
	m.SelectMocks[i].InPrompt = prompt
	m.SelectMocks[i].InSize = size
	m.SelectMocks[i].InItems = items
	m.SelectMocks[i].InSearch = search
	return m.SelectMocks[i].OutItem, m.SelectMocks[i].OutError
}
