package template

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/gardenbed/charm/ui"
)

var (
	templateFiles = []string{"template.yml", "template.yaml"}
)

// Template has all specifications for a Basil code template.
type Template struct {
	path string

	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Changes     Changes `yaml:"changes"`
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

// Read reads template specifications from a file in a path.
func Read(path string, params interface{}) (Template, error) {
	file, err := findFile(path)
	if err != nil {
		return Template{}, err
	}

	t, err := template.ParseFiles(file)
	if err != nil {
		return Template{}, err
	}

	buf := new(bytes.Buffer)
	template := Template{
		path: path,
	}

	if err := t.Execute(buf, params); err != nil {
		return Template{}, err
	}

	if err := yaml.NewDecoder(buf).Decode(&template); err != nil {
		return Template{}, err
	}

	return template, nil
}

// Changes define all the required changes for a template.
type Changes struct {
	Deletes  Deletes  `yaml:"deletes"`
	Moves    Moves    `yaml:"moves"`
	Appends  Appends  `yaml:"appends"`
	Replaces Replaces `yaml:"replaces"`
}

func (c *Changes) execute(root string, u ui.UI) error {
	if err := c.Deletes.execute(root, u); err != nil {
		return err
	}

	if err := c.Moves.execute(root, u); err != nil {
		return err
	}

	if err := c.Appends.execute(root, u); err != nil {
		return err
	}

	if err := c.Replaces.execute(root, u); err != nil {
		return err
	}

	return nil
}

// Delete is used for deteling files or directories.
type Delete struct {
	Glob string
}

// Deletes is the type for a slice of Delete type.
type Deletes []Delete

func (d Deletes) execute(root string, u ui.UI) error {
	for _, delete := range d {
		glob := filepath.Join(root, delete.Glob)
		matches, err := filepath.Glob(glob)
		if err != nil {
			return err
		}

		for _, match := range matches {
			u.Debugf(ui.Green, "Removing %s", match)
			if err := os.RemoveAll(match); err != nil {
				return err
			}
		}
	}

	return nil
}

// Move is used for moving (renaming) a file or directory.
type Move struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

// Moves is the type for a slice of Move type.
type Moves []Move

func (m Moves) execute(root string, u ui.UI) error {
	for _, move := range m {
		src := filepath.Join(root, move.Src)
		dest := filepath.Join(root, move.Dest)
		dir := filepath.Dir(dest)

		u.Debugf(ui.Green, "Creating directory %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		u.Debugf(ui.Green, "Moving %s to %s", src, dest)
		if err := os.Rename(src, dest); err != nil {
			return err
		}
	}

	return nil
}

// Append is used for adding content to files.
type Append struct {
	Filepath string `yaml:"filepath"`
	Content  string `yaml:"content"`
}

// Appends is the type for a slice of Append type.
type Appends []Append

func (a Appends) execute(root string, u ui.UI) error {
	for _, append := range a {
		path := filepath.Join(root, append.Filepath)

		u.Debugf(ui.Green, "Editing %s", path)
		u.Tracef(ui.Yellow, "Reading %s", path)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		u.Tracef(ui.Magenta, "  Appending %q", append.Content)
		if _, err := fmt.Fprintln(f, append.Content); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Replace is used for replacing content in files.
type Replace struct {
	Filepath string `yaml:"filepath"`
	Old      string `yaml:"old"`
	New      string `yaml:"new"`
}

// Replaces is the type for a slice of Replace type.
type Replaces []Replace

func (r Replaces) execute(root string, u ui.UI) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			var data []byte

			for _, replace := range r {
				re, err := regexp.Compile(replace.Filepath)
				if err != nil {
					return err
				}

				if re.MatchString(path) {
					if data == nil {
						u.Debugf(ui.Green, "Editing %s", path)
						u.Tracef(ui.Yellow, "Reading %s", path)
						if data, err = ioutil.ReadFile(path); err != nil {
							return err
						}
					}

					u.Tracef(ui.Magenta, "  Replacing %q with %q", replace.Old, replace.New)
					data = bytes.ReplaceAll(data, []byte(replace.Old), []byte(replace.New))
				}
			}

			if data != nil {
				u.Tracef(ui.Yellow, "Writing back %s", path)
				if err := ioutil.WriteFile(path, data, 0); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
