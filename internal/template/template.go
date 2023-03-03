package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gardenbed/charm/ui"
)

// Params define required params for a template.
type Params []string

func (p Params) Has(param string) bool {
	for _, val := range p {
		if param == val {
			return true
		}
	}

	return false
}

// Template has all specifications for a Basil code template.
type Template struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Edits       Edits  `yaml:"edits"`
}

// Execute runs all the edits defined for the template.
func (t *Template) Execute(u ui.UI, root string) error {
	return t.Edits.execute(u, root)
}

// Edits define all the required edits for a template.
type Edits struct {
	Deletes  Deletes  `yaml:"deletes"`
	Moves    Moves    `yaml:"moves"`
	Appends  Appends  `yaml:"appends"`
	Replaces Replaces `yaml:"replaces"`
}

func (e *Edits) execute(u ui.UI, root string) error {
	if err := e.Deletes.execute(u, root); err != nil {
		return err
	}

	if err := e.Moves.execute(u, root); err != nil {
		return err
	}

	if err := e.Appends.execute(u, root); err != nil {
		return err
	}

	if err := e.Replaces.execute(u, root); err != nil {
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

func (d Deletes) execute(u ui.UI, root string) error {
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

func (m Moves) execute(u ui.UI, root string) error {
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

func (a Appends) execute(u ui.UI, root string) error {
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

func (r Replaces) execute(u ui.UI, root string) error {
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
						if data, err = os.ReadFile(path); err != nil {
							return err
						}
					}

					u.Tracef(ui.Magenta, "  Replacing %q with %q", replace.Old, replace.New)
					data = bytes.ReplaceAll(data, []byte(replace.Old), []byte(replace.New))
				}
			}

			if data != nil {
				u.Tracef(ui.Yellow, "Writing back %s", path)
				if err := os.WriteFile(path, data, 0); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
