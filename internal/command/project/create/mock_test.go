package create

import (
	"context"
	"io"

	"github.com/gardenbed/go-github"

	"github.com/gardenbed/basil-cli/internal/archive"
	"github.com/gardenbed/basil-cli/internal/template"
	"github.com/gardenbed/basil-cli/internal/ui"
)

type (
	AskMock struct {
		InPrompt   string
		InDefault  string
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

		AskIndex int
		AskMocks []AskMock

		SelectIndex int
		SelectMocks []SelectMock
	}
)

func (m *MockUI) Ask(prompt, Default string, validate ui.ValidateFunc) (string, error) {
	i := m.AskIndex
	m.AskIndex++
	m.AskMocks[i].InPrompt = prompt
	m.AskMocks[i].InDefault = Default
	m.AskMocks[i].InValidate = validate
	return m.AskMocks[i].OutValue, m.AskMocks[i].OutError
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

type (
	DownloadTarArchiveMock struct {
		InContext   context.Context
		InRef       string
		InWriter    io.Writer
		OutResponse *github.Response
		OutError    error
	}

	MockRepoService struct {
		DownloadTarArchiveIndex int
		DownloadTarArchiveMocks []DownloadTarArchiveMock
	}
)

func (m *MockRepoService) DownloadTarArchive(ctx context.Context, ref string, writer io.Writer) (*github.Response, error) {
	i := m.DownloadTarArchiveIndex
	m.DownloadTarArchiveIndex++
	m.DownloadTarArchiveMocks[i].InContext = ctx
	m.DownloadTarArchiveMocks[i].InRef = ref
	m.DownloadTarArchiveMocks[i].InWriter = writer
	return m.DownloadTarArchiveMocks[i].OutResponse, m.DownloadTarArchiveMocks[i].OutError
}

type (
	ExtractMock struct {
		InDest     string
		InReader   io.Reader
		InSelector archive.Selector
		OutError   error
	}

	MockArchiveService struct {
		ExtractIndex int
		ExtractMocks []ExtractMock
	}
)

func (m *MockArchiveService) Extract(dest string, reader io.Reader, selector archive.Selector) error {
	i := m.ExtractIndex
	m.ExtractIndex++
	m.ExtractMocks[i].InDest = dest
	m.ExtractMocks[i].InReader = reader
	m.ExtractMocks[i].InSelector = selector
	return m.ExtractMocks[i].OutError
}

type (
	ExecuteMock struct {
		InTemplate template.Template
		OutError   error
	}

	MockTemplateService struct {
		ExecuteIndex int
		ExecuteMocks []ExecuteMock
	}
)

func (m *MockTemplateService) Execute(template template.Template) error {
	i := m.ExecuteIndex
	m.ExecuteIndex++
	m.ExecuteMocks[i].InTemplate = template
	return m.ExecuteMocks[i].OutError
}
