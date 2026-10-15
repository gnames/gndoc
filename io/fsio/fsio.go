// Package fsio provides access to the file system and the user's
// environment.
package fsio

import (
	"fmt"
	"os"

	"github.com/gnames/gnsys"
)

// ReadFile returns the content of a file. It returns an error if the
// file does not exist.
func ReadFile(path string) ([]byte, error) {
	exists, err := gnsys.FileExists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("file '%s' does not exist", path)
	}
	return os.ReadFile(path)
}

// HomeDir returns the home directory of the user.
func HomeDir() (string, error) {
	return os.UserHomeDir()
}
