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

	MockUI struct {
		ui.UI

		AskIndex int
		AskMocks []AskMock
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
	LoadMock struct {
		InPath   string
		OutError error
	}

	ParamsMock struct {
		OutParams template.Params
	}

	TemplateMock struct {
		InInputs    interface{}
		OutTemplate *template.Template
		OutError    error
	}

	MockTemplateService struct {
		LoadIndex int
		LoadMocks []LoadMock

		ParamsIndex int
		ParamsMocks []ParamsMock

		TemplateIndex int
		TemplateMocks []TemplateMock
	}
)

func (m *MockTemplateService) Load(path string) error {
	i := m.LoadIndex
	m.LoadIndex++
	m.LoadMocks[i].InPath = path
	return m.LoadMocks[i].OutError
}

func (m *MockTemplateService) Params() template.Params {
	i := m.ParamsIndex
	m.ParamsIndex++
	return m.ParamsMocks[i].OutParams
}

func (m *MockTemplateService) Template(inputs interface{}) (*template.Template, error) {
	i := m.TemplateIndex
	m.TemplateIndex++
	m.TemplateMocks[i].InInputs = inputs
	return m.TemplateMocks[i].OutTemplate, m.TemplateMocks[i].OutError
}
