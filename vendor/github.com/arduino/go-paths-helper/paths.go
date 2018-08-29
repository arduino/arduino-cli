/*
 * This file is part of PathsHelper library.
 *
 * Copyright 2018 Arduino AG (http://www.arduino.cc/)
 *
 * PathsHelper library is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 */

package paths

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Path represents a path
type Path struct {
	path string
}

// New creates a new Path object. If path is the empty string
// then nil is returned.
func New(path ...string) *Path {
	if len(path) == 0 {
		return New("")
	}
	res := &Path{path: path[0]}
	if len(path) > 1 {
		return res.Join(path[1:]...)
	}
	return res
}

// NewFromFile creates a new Path object using the path name
// obtained from the File object (see os.File.Name function).
func NewFromFile(file *os.File) *Path {
	return New(file.Name())
}

// Stat returns a FileInfo describing the named file. The result is
// cached internally for next queries. To ensure that the cached
// FileInfo entry is updated just call Stat again.
func (p *Path) Stat() (os.FileInfo, error) {
	return os.Stat(p.path)
}

// Clone create a copy of the Path object
func (p *Path) Clone() *Path {
	return New(p.path)
}

// Join create a new Path by joining the provided paths
func (p *Path) Join(paths ...string) *Path {
	return New(filepath.Join(p.path, filepath.Join(paths...)))
}

// JoinPath create a new Path by joining the provided paths
func (p *Path) JoinPath(paths ...*Path) *Path {
	res := p.Clone()
	for _, path := range paths {
		res = res.Join(path.path)
	}
	return res
}

// Base returns the last element of path
func (p *Path) Base() string {
	return filepath.Base(p.path)
}

// Ext returns the file name extension used by path
func (p *Path) Ext() string {
	return filepath.Ext(p.path)
}

// HasPrefix returns true if the file name has one of the
// given prefixes (the Base() method is used to obtain the
// file name used for the comparison)
func (p *Path) HasPrefix(prefixes ...string) bool {
	filename := p.Base()
	for _, prefix := range prefixes {
		if strings.HasPrefix(filename, prefix) {
			return true
		}
	}
	return false
}

// HasSuffix returns true if the file name has one of the
// given suffixies
func (p *Path) HasSuffix(suffixies ...string) bool {
	filename := p.String()
	for _, suffix := range suffixies {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}

// RelTo returns a relative Path that is lexically equivalent to r when
// joined to the current Path
func (p *Path) RelTo(r *Path) (*Path, error) {
	rel, err := filepath.Rel(p.path, r.path)
	if err != nil {
		return nil, err
	}
	return New(rel), nil
}

// Abs returns the absolute path of the current Path
func (p *Path) Abs() (*Path, error) {
	abs, err := filepath.Abs(p.path)
	if err != nil {
		return nil, err
	}
	return New(abs), nil
}

// IsAbs returns true if the Path is absolute
func (p *Path) IsAbs() bool {
	return filepath.IsAbs(p.path)
}

// ToAbs transofrm the current Path to the corresponding absolute path
func (p *Path) ToAbs() error {
	abs, err := filepath.Abs(p.path)
	if err != nil {
		return err
	}
	p.path = abs
	return nil
}

// Clean Clean returns the shortest path name equivalent to path by
// purely lexical processing
func (p *Path) Clean() *Path {
	return New(filepath.Clean(p.path))
}

// IsInsideDir returns true if the current path is inside the provided
// dir
func (p *Path) IsInsideDir(dir *Path) (bool, error) {
	rel, err := filepath.Rel(p.path, dir.path)
	if err != nil {
		return false, err
	}
	return strings.Contains(rel, ".."+string(os.PathSeparator)) || rel == "..", nil
}

// Parent returns all but the last element of path, typically the path's
// directory or the parent directory if the path is already a directory
func (p *Path) Parent() *Path {
	return New(filepath.Dir(p.path))
}

// Mkdir create a directory denoted by the current path
func (p *Path) Mkdir() error {
	return os.Mkdir(p.path, 0755)
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error
func (p *Path) MkdirAll() error {
	return os.MkdirAll(p.path, os.FileMode(0755))
}

// Remove removes the named file or directory
func (p *Path) Remove() error {
	return os.Remove(p.path)
}

// RemoveAll removes path and any children it contains. It removes
// everything it can but returns the first error it encounters. If
// the path does not exist, RemoveAll returns nil (no error).
func (p *Path) RemoveAll() error {
	return os.RemoveAll(p.path)
}

// Rename renames (moves) the path to newpath. If newpath already exists
// and is not a directory, Rename replaces it. OS-specific restrictions
// may apply when oldpath and newpath are in different directories. If
// there is an error, it will be of type *os.LinkError.
func (p *Path) Rename(newpath *Path) error {
	return os.Rename(p.path, newpath.path)
}

// MkTempDir creates a new temporary directory inside the path
// pointed by the Path object with a name beginning with prefix
// and returns the path of the new directory.
func (p *Path) MkTempDir(prefix string) (*Path, error) {
	return MkTempDir(p.path, prefix)
}

// FollowSymLink transforms the current path to the path pointed by the
// symlink if path is a symlink, otherwise it does nothing
func (p *Path) FollowSymLink() error {
	resolvedPath, err := filepath.EvalSymlinks(p.path)
	if err != nil {
		return err
	}
	p.path = resolvedPath
	return nil
}

// Exist return true if the path exists
func (p *Path) Exist() (bool, error) {
	_, err := p.Stat()
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsDir return true if the path exists and is a directory
func (p *Path) IsDir() (bool, error) {
	info, err := p.Stat()
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ReadDir returns a PathList containing the content of the directory
// pointed by the current Path
func (p *Path) ReadDir() (PathList, error) {
	infos, err := ioutil.ReadDir(p.path)
	if err != nil {
		return nil, err
	}
	paths := PathList{}
	for _, info := range infos {
		path := p.Clone().Join(info.Name())
		paths.Add(path)
	}
	return paths, nil
}

// CopyTo copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func (p *Path) CopyTo(dst *Path) error {
	in, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst.path)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	si, err := p.Stat()
	if err != nil {
		return err
	}

	err = os.Chmod(dst.path, si.Mode())
	if err != nil {
		return err
	}

	return nil
}

// CopyDirTo recursively copies the directory denoted by the current path to
// the destination path. The source directory must exist and the destination
// directory must NOT exist (no implicit destination name allowed).
// Symlinks are not copied, they will be supported in future versions.
func (p *Path) CopyDirTo(dst *Path) error {
	src := p.Clean()
	dst = dst.Clean()

	srcInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("getting stat infor for %s: %s", src, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	if exist, err := dst.Exist(); exist {
		return fmt.Errorf("destination %s already exists", dst)
	} else if err != nil {
		return fmt.Errorf("checking if %s exists: %s", dst, err)
	}

	if err := dst.MkdirAll(); err != nil {
		return fmt.Errorf("creating destination dir %s: %s", dst, err)
	}
	if err := os.Chmod(dst.path, srcInfo.Mode()); err != nil {
		return fmt.Errorf("setting permission for dir %s: %s", dst, err)
	}

	srcFiles, err := src.ReadDir()
	if err != nil {
		return fmt.Errorf("error reading source dir %s: %s", src, err)
	}

	for _, srcPath := range srcFiles {
		srcPathInfo, err := srcPath.Stat()
		if err != nil {
			return fmt.Errorf("getting stat info for %s: %s", srcPath, err)
		}
		dstPath := dst.Join(srcPath.Base())

		if srcPathInfo.IsDir() {
			if err := srcPath.CopyDirTo(dstPath); err != nil {
				return fmt.Errorf("copying %s to %s: %s", srcPath, dstPath, err)
			}
			continue
		}

		// Skip symlinks.
		if srcPathInfo.Mode()&os.ModeSymlink != 0 {
			// TODO
			continue
		}

		if err := srcPath.CopyTo(dstPath); err != nil {
			return fmt.Errorf("copying %s to %s: %s", srcPath, dstPath, err)
		}
	}
	return nil
}

// Chtimes changes the access and modification times of the named file,
// similar to the Unix utime() or utimes() functions.
func (p *Path) Chtimes(atime, mtime time.Time) error {
	return os.Chtimes(p.path, atime, mtime)
}

// ReadFile reads the file named by filename and returns the contents
func (p *Path) ReadFile() ([]byte, error) {
	return ioutil.ReadFile(p.path)
}

// WriteFile writes data to a file named by filename. If the file
// does not exist, WriteFile creates it otherwise WriteFile truncates
// it before writing.
func (p *Path) WriteFile(data []byte) error {
	return ioutil.WriteFile(p.path, data, os.FileMode(0644))
}

// ReadFileAsLines reads the file named by filename and returns it as an
// array of lines. This function takes care of the newline encoding
// differences between different OS
func (p *Path) ReadFileAsLines() ([]string, error) {
	data, err := p.ReadFile()
	if err != nil {
		return nil, err
	}
	txt := string(data)
	txt = strings.Replace(txt, "\r\n", "\n", -1)
	return strings.Split(txt, "\n"), nil
}

// Truncate create an empty file named by path or if the file already
// exist it trucates it (delete all contents)
func (p *Path) Truncate() error {
	return p.WriteFile([]byte{})
}

// EqualsTo return true if both paths are equal
func (p *Path) EqualsTo(other *Path) bool {
	return p.path == other.path
}

// EquivalentTo return true if both paths are equivalent (they points to the
// same file even if they are lexicographically different)
func (p *Path) EquivalentTo(other *Path) bool {
	return p.Clean().path == other.Clean().path
}

func (p *Path) String() string {
	return p.path
}
