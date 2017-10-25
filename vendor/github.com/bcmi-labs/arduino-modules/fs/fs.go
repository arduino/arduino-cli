// Package fs is an abstraction of the filesystem. In our apps we often have to
// perform operations both on the filesystem and on s3, and we need this operation
// to be completely transparent to the library. Also it doesn't hurt to have something
// ready to be marshalled into json or rethinkdb.
package fs

import (
	"os"
	"path"
	"time"

	"github.com/juju/errors"
)

// File represents a single file in the filesystem.
// It is ready to be marshalled as json or as a rethinkdb document.
// As rethinkdb document the data will be hidden.
type File struct {
	Path         string    `json:"path" gorethink:"path"`
	Data         []byte    `json:"data,omitempty" gorethink:"-"`
	Name         string    `json:"name,omitempty" gorethink:"name"`
	Mime         string    `json:"mimetype,omitempty" gorethink:"mime"`
	Size         int64     `json:"size,omitempty" gorethink:"size"`
	LastModified time.Time `json:"last_modified,omitempty" gorethink:"last_modified"`
}

// Reader allows to read the content of a file on a filesystem
type Reader interface {
	ReadFile(filename string) ([]byte, error)
}

// Read populates the Data property of the file by using the fs param
func (f *File) Read(fs Reader) error {
	var err error
	f.Data, err = fs.ReadFile(f.Path)
	if err != nil {
		return err
	}
	f.Size = int64(len(f.Data))
	return nil
}

// Writer allows to write the content of a file on a filesystem
type Writer interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
}

// Write persists the data property on the fs, with the permission given
func (f *File) Write(fs Writer, perm os.FileMode) error {
	err := fs.MkdirAll(path.Dir(f.Path), perm)
	if err != nil {
		return errors.Annotatef(err, "while creating the directory %s", path.Dir(f.Path))
	}

	err = fs.WriteFile(f.Path, f.Data, perm)
	if err != nil {
		return errors.Annotatef(err, "while writing %s with data %s", f.Path, f.Data)
	}
	return nil
}

// Remover allows to write the content of a file on a filesystem
type Remover interface {
	Remove(name string) error
}

// Lister allows to get a list of files from a prefix
type Lister interface {
	List(prefix string) ([]File, error)
}

// Manager combines all the interfaces
type Manager interface {
	Reader
	Writer
	Remover
	Lister
}
