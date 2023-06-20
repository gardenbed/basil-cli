package template

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gardenbed/charm/ui"
	"gopkg.in/yaml.v3"
)

var (
	templateFiles = []string{"template.yml", "template.yaml"}
	paramRegexp   = regexp.MustCompile(`{{\s*\.([A-Z][0-9A-Za-z]+)\s*}}`)
)

// Service
type Service struct {
	ui     ui.UI
	path   string
	text   string
	params Params
}

// NewService creates a new template service.
func NewService(ui ui.UI) *Service {
	return &Service{
		ui: ui,
	}
}

func findFile(path string) (string, error) {
	for _, templateFile := range templateFiles {
		file := filepath.Join(path, templateFile)
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}

		return file, nil
	}

	return "", errors.New("template file not found")
}

// Load reads a template YAML file and makes it available for other methods.
func (s *Service) Load(path string) error {
	filename, err := findFile(path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	s.path = path
	s.text = string(data)

	for _, match := range paramRegexp.FindAllStringSubmatch(s.text, -1) {
		if len(match) == 2 {
			s.params = append(s.params, match[1])
		}
	}

	return nil
}

func (s *Service) Params() Params {
	return s.params
}

func (s *Service) Template(inputs interface{}) (*Template, error) {
	t, err := template.New("yaml").Parse(s.text)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, inputs); err != nil {
		return nil, err
	}

	template := new(Template)
	if err := yaml.NewDecoder(buf).Decode(template); err != nil {
		return nil, err
	}

	return template, nil
}
