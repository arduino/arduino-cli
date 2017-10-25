package iohelpers

import "os"

// Exists returns whether the given file or directory exists or not
// From http://stackoverflow.com/a/10510783
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
