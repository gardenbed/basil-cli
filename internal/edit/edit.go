package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gardenbed/basil-cli/internal/debug"
)

// Editor is used for editing text files.
type Editor struct {
	debugger *debug.DebuggerSet
}

// NewEditor creates a new editor.
func NewEditor(level debug.Level) *Editor {
	return &Editor{
		debugger: debug.NewSet(level),
	}
}

// Remove deletes files and folders using glob patterrns.
func (e *Editor) Remove(globs ...string) error {
	for _, glob := range globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return err
		}

		for _, match := range matches {
			if err := os.RemoveAll(match); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveSpec has the input parameters for the Move method.
type MoveSpec struct {
	Src  string
	Dest string
}

// Move moves files from a destination to a source.
// If mkdir is true, the destination path will be created.
func (e *Editor) Move(mkdir bool, specs ...MoveSpec) error {
	for _, s := range specs {
		if mkdir {
			dir := filepath.Dir(s.Dest)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}

		if err := os.Rename(s.Src, s.Dest); err != nil {
			return err
		}
	}

	return nil
}

// AppendSpec has the input parameters for the Append method.
type AppendSpec struct {
	Path    string
	Content string
}

// Append appends new lines to files.
// If create is true, the target file will be created if it does not exist.
func (e *Editor) Append(create bool, specs ...AppendSpec) error {
	for _, s := range specs {
		flag := os.O_WRONLY
		if create {
			flag |= os.O_CREATE
		}

		f, err := os.OpenFile(s.Path, flag, 0644)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(f, s.Content); err != nil {
			return err
		}
	}

	return nil
}

// ReplaceSpec has the input parameters for the ReplaceInDir method.
type ReplaceSpec struct {
	PathRE *regexp.Regexp
	OldRE  *regexp.Regexp
	New    string
}

// ReplaceInDir is used for modifying all files in a directory.
func (e *Editor) ReplaceInDir(root string, specs ...ReplaceSpec) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			var data []byte

			for _, s := range specs {
				if s.PathRE.MatchString(path) {
					if data == nil {
						e.debugger.Yellow.Tracef("Reading %s", path)
						e.debugger.Green.Debugf("Editing %s", path)
						if data, err = ioutil.ReadFile(path); err != nil {
							return err
						}
					}

					e.debugger.Magenta.Tracef("  Replacing %q with %q", s.OldRE, s.New)
					data = s.OldRE.ReplaceAll(data, []byte(s.New))
				}
			}

			if data != nil {
				e.debugger.Yellow.Tracef("Writing back %s", path)
				if err := ioutil.WriteFile(path, data, 0); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
