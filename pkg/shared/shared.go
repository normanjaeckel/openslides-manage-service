package shared

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
)

// CreateFile creates a file in the given directory with the given content.
// Use a truthy value for force to override an existing file.
func CreateFile(dir string, force bool, name string, content []byte, secure bool) error {
	p := path.Join(dir, name)

	pExists, err := fileExists(p)
	if err != nil {
		return fmt.Errorf("checking file existance: %w", err)
	}
	if !force && pExists {
		// No force-mode and file already exists, so skip this file.
		return nil
	}

	var perm fs.FileMode
	perm = 0666
	if secure {
		perm = 0660
	}
	if err := os.WriteFile(p, content, perm); err != nil {
		return fmt.Errorf("creating and writing to file %q: %w", p, err)
	}
	return nil
}

// fileExists is a small helper function to check if a file already exists. It is not
// save in concurrent usage.
func fileExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("checking existance of file %s: %w", p, err)
}