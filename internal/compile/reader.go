package compile

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// getModuleName returns the name of go module from a given path.
func getModuleName(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(path, "go.mod")

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			if parent := filepath.Dir(path); parent != "/" {
				fmt.Println(parent)
				return getModuleName(parent)
			}
		}
		return "", err
	}

	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("invalid go.mod file: no module name found")
}

type visitFunc func(baseDir, relDir string) error

// readPackages visits all package directories recursively in a given path.
func readPackages(path string, visit visitFunc) error {
	return packageDirs(path, ".", visit)
}

func packageDirs(baseDir, relDir string, visit visitFunc) error {
	if err := visit(baseDir, relDir); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(filepath.Join(baseDir, relDir))
	if err != nil {
		return err
	}

	for _, file := range files {
		// Any directory that starts with "." is NOT considered
		if file.IsDir() && isPackageDir(file.Name()) {
			subdir := filepath.Join(relDir, file.Name())
			if err := packageDirs(baseDir, subdir, visit); err != nil {
				return err
			}
		}
	}

	return nil
}

func isPackageDir(name string) bool {
	startsWithDot := strings.HasPrefix(name, ".")
	return !startsWithDot && name != "bin"
}
