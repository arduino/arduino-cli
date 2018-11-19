package fs

import (
	"context"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/gofrs/flock"
	"github.com/juju/errors"
)

// HashDisk is a disk that spreads files in directories according to their name in order not to have a single folder with millions of subdirectories.
// For example the filename test/banana will be spread in te/st/test/banana
// Normally you would end up with an unbalanced distribution of folders, but since in real application filenames are already hashed (eg 6bf5ee0aea544986e84bf7abd5c96bde:matteosuppo/sketches/f7aca8a9-29f4-4420-9c7b-61d4e4064b84/sketch_sep25a/ReadMe.adoc), the problem is negligible
type HashDisk struct {
	Base             string
	NameOverride     string
	CharacterMapping map[string]string
}

func (d *HashDisk) Name() string {
	if d.NameOverride != "" {
		return d.NameOverride
	}
	return "HashDisk"
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
// It returns a notfound error if the file is missing
func (d *HashDisk) ReadFile(filename string) ([]byte, error) {
	hash := strings.Replace(filename, string(filepath.Separator), "", -1)
	filename = filepath.Join(hash[0:2], hash[2:4], filename)

	filename = d.replaceCharacters(filename)
	filename = filepath.Join(d.Base, filename)

	_, err := os.Stat(filename)
	if err != nil {
		return nil, errors.NewNotFound(err, filename)
	}

	fileLock, err := lock(filename)
	if err != nil {
		return nil, errors.Annotatef(err, "trylock %s", filename)
	}
	defer unlock(fileLock)

	data, err := ioutil.ReadFile(filename)
	if err != nil && os.IsNotExist(err) {
		return nil, errors.NewNotFound(err, filename)
	}

	return data, err
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func (d *HashDisk) WriteFile(filename string, data []byte, perm os.FileMode) error {
	hash := strings.Replace(filename, string(filepath.Separator), "", -1)
	filename = filepath.Join(hash[0:2], hash[2:4], filename)

	filename = d.replaceCharacters(filename)
	filename = filepath.Join(d.Base, filename)

	fileLock, err := rlock(filename)
	if err != nil {
		return errors.Annotatef(err, "trylock %s", filename)
	}

	defer unlock(fileLock)

	err = ioutil.WriteFile(filename, data, perm)
	if err != nil {
		return errors.Annotatef(err, "write %s", filename)
	}

	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error.
// The permission bits perm are used for all directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing and returns nil.
func (d *HashDisk) MkdirAll(path string, perm os.FileMode) error {
	hash := strings.Replace(path, string(filepath.Separator), "", -1)
	path = filepath.Join(hash[0:2], hash[2:4], path)

	path = d.replaceCharacters(path)
	path = filepath.Join(d.Base, path)
	return os.MkdirAll(path, perm)
}

// Remove removes the named file or directory (children included).
// It fails silently if the file doesn't exist
func (d *HashDisk) Remove(name string) error {
	hash := strings.Replace(name, string(filepath.Separator), "", -1)
	name = filepath.Join(hash[0:2], hash[2:4], name)

	name = d.replaceCharacters(name)
	name = filepath.Join(d.Base, name)

	stat, err := os.Stat(name)
	if err != nil {
		return nil // If there's no file, there's no reason to fail
	}

	lockfile := name
	if stat.IsDir() {
		lockfile = lockfile + ".lock"
	}

	fileLock, err := lock(lockfile)
	if err != nil {
		return errors.Annotatef(err, "trylock %s", name)
	}
	defer unlock(fileLock)
	defer os.RemoveAll(lockfile)

	err = os.RemoveAll(name)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// List returns a list of files in a directory
func (d *HashDisk) List(prefix string) ([]File, error) {
	hash := strings.Replace(prefix, string(filepath.Separator), "", -1)
	prefix = filepath.Join(hash[0:2], hash[2:4], prefix)

	prefix = d.replaceCharacters(prefix)
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
			Path: strings.Replace(path, d.Base+"/", "", 1)[6:],
			Name: filepath.Base(path),
			Size: f.Size(),
		})

		return nil
	})
	return list, err
}

// replaceCharacters replaces the characters in the CharacterMapping
// from the filename/path given
func (d *HashDisk) replaceCharacters(filename string) string {
	if d.CharacterMapping != nil {
		for k, v := range d.CharacterMapping {
			filename = strings.Replace(filename, k, v, -1)
		}
	}
	return filename
}

// ReadDir returns a list of file contained under a path
func (d *HashDisk) ReadDir(path string) ([]File, error) {

	hash := strings.Replace(path, string(filepath.Separator), "", -1)
	path = filepath.Join(hash[0:2], hash[2:4], path)

	path = d.replaceCharacters(path)
	path = filepath.Join(d.Base, path)

	iofiles, err := ioutil.ReadDir(path)
	if err != nil && os.IsNotExist(err) {
		return nil, errors.NewNotFound(err, path)
	}

	var files []File

	for _, f := range iofiles {
		file := File{
			Path:         strings.Replace(path, d.Base+"/", "", 1)[6:] + "/" + f.Name(),
			Name:         f.Name(),
			Size:         f.Size(),
			LastModified: f.ModTime(),
			IsDir:        f.IsDir(),
			Mime:         mime.TypeByExtension(filepath.Ext(f.Name())),
		}
		files = append(files, file)
	}
	return files, err
}

// lock attempts to lock the file for 10 seconds. If 10 seconds pass without success it returns an error
func lock(filename string) (*flock.Flock, error) {
	fileLock := flock.NewFlock(filename)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}

	if !locked {
		return nil, errors.New("could not lock")
	}

	return fileLock, nil
}

// rlock attempts to gain a shared lock for the given file for 10 seconds. If 10 seconds pass without success it returns an error
func rlock(filename string) (*flock.Flock, error) {
	fileLock := flock.NewFlock(filename)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	locked, err := fileLock.TryRLockContext(ctx, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}

	if !locked {
		return nil, errors.New("could not rlock")
	}

	return fileLock, nil
}

// unlock panics so that the error is not swallowed by the defer
// if it panics it means something is very wrong indeed anyway
func unlock(lock *flock.Flock) {
	err := lock.Unlock()
	if err != nil {
		panic(err)
	}
}
