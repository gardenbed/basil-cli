package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gardenbed/charm/ui"
)

// Selector is a function for selecting and projecting a directory or file.
// If a selector function returns true, the directory or file will be extracted.
// The selector function can also maps the give path to a new path when extracting.
type Selector func(string) (string, bool)

// TarArchive facilitates working with tar.gz files.
type TarArchive struct {
	ui ui.UI
}

// NewTarArchive creates a new instance of TarArchive.
func NewTarArchive(ui ui.UI) *TarArchive {
	return &TarArchive{
		ui: ui,
	}
}

// Extract traverses all directories and files in a tar.gz archive and uses a selector function to extract directories and files.
// dest is the path for extracing the directories and files into.
func (a *TarArchive) Extract(dest string, r io.Reader, f Selector) error {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("error on creating gzip reader: %s", err)
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error on reading tar reader: %s", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if path, ok := f(header.Name); ok {
				path = filepath.Join(dest, path)
				if err := os.Mkdir(path, 0755); err != nil {
					return fmt.Errorf("error on creating directory: %s", err)
				}
				a.ui.Debugf(ui.Cyan, "Directory created: %s", path)
			}

		case tar.TypeReg:
			if path, ok := f(header.Name); ok {
				path = filepath.Join(dest, path)
				file, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("error on creating file: %s", err)
				}

				if _, err := io.Copy(file, tarReader); err != nil {
					return fmt.Errorf("error on copying from tar reader: %s", err)
				}

				a.ui.Debugf(ui.Cyan, "  File copied: %s", path)
			}

		case tar.TypeXGlobalHeader:
			// Ignore the pax global header in GitHub generated tarballs

		default:
			return fmt.Errorf("%s: unknown header type: %c", header.Name, header.Typeflag)
		}
	}

	return nil
}
