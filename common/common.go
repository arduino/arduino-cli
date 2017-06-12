package common

import (
	"fmt"
	"os/user"
	"path/filepath"
	"runtime"
)

// GetDefaultArduinoFolder returns the default data folder for Arduino platform
func GetDefaultArduinoFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(usr.HomeDir, ".arduino15"), nil
	case "darwin":
		return filepath.Join(usr.HomeDir, "Library", "arduino15"), nil
	default:
		return "", fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
	}
}

// GetDefaultLibFolder get the default folder of downloaded libraries.
func GetDefaultLibFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, "Arduino", "libraries"), nil
}
