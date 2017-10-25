package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juju/errors"
)

// Disk is a wrapper around the usual os and ioutil golang libraries
type Disk struct {
	Base string
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
// It returns a notfound error if the file is missing
func (d *Disk) ReadFile(filename string) ([]byte, error) {
	filename = filepath.Join(d.Base, filename)

	data, err := ioutil.ReadFile(filename)
	if err != nil && os.IsNotExist(err) {
		return nil, errors.NewNotFound(err, filename)
	}

	return data, err
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func (d *Disk) WriteFile(filename string, data []byte, perm os.FileMode) error {
	filename = filepath.Join(d.Base, filename)
	return ioutil.WriteFile(filename, data, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error.
// The permission bits perm are used for all directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing and returns nil.
func (d *Disk) MkdirAll(path string, perm os.FileMode) error {
	path = filepath.Join(d.Base, path)
	return os.MkdirAll(path, perm)
}

// Remove removes the named file or directory. It fails silently if the file didn't exist
func (d *Disk) Remove(name string) error {
	name = filepath.Join(d.Base, name)
	err := os.Remove(name)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// List returns a list of files in a directory
func (d *Disk) List(prefix string) ([]File, error) {
	prefix = filepath.Join(d.Base, prefix)

	list := []File{}

	err := filepath.Walk(prefix, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if f.IsDir() {
			return nil
		}

		list = append(list, File{
			Name: filepath.Base(path),
			Size: f.Size(),
		})

		return nil
	})
	return list, err
}
