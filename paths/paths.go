package paths

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Path represents a path
type Path struct {
	path               string
	cachedFileInfo     os.FileInfo
	cachedFileInfoTime time.Time
}

// New creates a new Path object
func New(path string) *Path {
	return &Path{path: path}
}

func (p *Path) setCachedFileInfo(info os.FileInfo) {
	p.cachedFileInfo = info
	p.cachedFileInfoTime = time.Now()
}

// Stat returns a FileInfo describing the named file. The result is
// cached internally for next queries. To ensure that the cached
// FileInfo entry is updated just call Stat again.
func (p *Path) Stat() (os.FileInfo, error) {
	info, err := os.Stat(p.path)
	if err != nil {
		return nil, err
	}
	p.setCachedFileInfo(info)
	return info, nil
}

func (p *Path) stat() (os.FileInfo, error) {
	if p.cachedFileInfo != nil {
		if p.cachedFileInfoTime.Add(50 * time.Millisecond).After(time.Now()) {
			return p.cachedFileInfo, nil
		}
	}
	return p.Stat()
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

// Base Base returns the last element of path
func (p *Path) Base() string {
	return filepath.Base(p.path)
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

// ToAbs transofrm the current Path to the corresponding absolute path
func (p *Path) ToAbs() error {
	abs, err := filepath.Abs(p.path)
	if err != nil {
		return err
	}
	p.path = abs
	return nil
}

// FollowSymLink transforms the current path to the path pointed by the
// symlink if path is a symlink, otherwise it does nothing
func (p *Path) FollowSymLink() error {
	resolvedPath, err := filepath.EvalSymlinks(p.path)
	if err != nil {
		return err
	}
	p.path = resolvedPath
	p.cachedFileInfo = nil
	return nil
}

// Exist return true if the path exists
func (p *Path) Exist() (bool, error) {
	_, err := p.stat()
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
	info, err := p.stat()
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
		path.setCachedFileInfo(info)
		paths.Add(path)
	}
	return paths, nil
}

func (p *Path) String() string {
	return p.path
}
