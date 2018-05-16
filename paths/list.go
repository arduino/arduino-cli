package paths

import (
	"strings"
)

// PathList is a list of Path
type PathList []*Path

// NewPathList creates a new PathList with the given paths
func NewPathList(paths ...string) PathList {
	res := PathList{}
	for _, path := range paths {
		res = append(res, New(path))
	}
	return res
}

// Clone returns a copy of the current PathList
func (p *PathList) Clone() PathList {
	res := PathList{}
	for _, path := range *p {
		res.Add(path.Clone())
	}
	return res
}

// AsStrings return this path list as a string array
func (p *PathList) AsStrings() []string {
	res := []string{}
	for _, path := range *p {
		res = append(res, path.String())
	}
	return res
}

// FilterDirs remove all entries except directories
func (p *PathList) FilterDirs() {
	res := (*p)[:0]
	for _, path := range *p {
		if isDir, _ := path.IsDir(); isDir {
			res = append(res, path)
		}
	}
	*p = res
}

// FilterOutHiddenFiles remove all hidden files (files with the name
// starting with ".")
func (p *PathList) FilterOutHiddenFiles() {
	p.FilterOutPrefix(".")
}

// FilterOutPrefix remove all entries having the specified prefix
func (p *PathList) FilterOutPrefix(prefix string) {
	res := (*p)[:0]
	for _, path := range *p {
		if !strings.HasPrefix(path.Base(), prefix) {
			res = append(res, path)
		}
	}
	*p = res
}

// Add adds a Path to the PathList
func (p *PathList) Add(path *Path) {
	*p = append(*p, path)
}
