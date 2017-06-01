package libraries

import "github.com/pmylund/sortutil"

// StatusContext keeps the current status of the libraries in the system
// (the list of libraries, revisions, installed paths, etc.)
type StatusContext struct {
	Libraries map[string]*Library
}

// Library represents a library in the system
type Library struct {
	Name          string
	Author        string
	Maintainer    string
	Sentence      string
	Paragraph     string
	Website       string
	Category      string
	Architectures []string
	Types         []string
	Releases      []*Release
	Installed     *Release
	Latest        *Release
}

// Release represents a release of a library
type Release struct {
	Version         string
	URL             string
	ArchiveFileName string
	Size            int
	Checksum        string
}

// Versions returns an array of all versions available of the library
func (l *Library) Versions() []string {
	res := make([]string, len(l.Releases))
	i := 0
	for _, r := range l.Releases {
		res[i] = r.Version
		i++
	}
	sortutil.CiAsc(res)
	return res
}

// AddLibrary adds an IndexRelease to the status context
func (l *StatusContext) AddLibrary(indexLib *IndexRelease) {
	name := indexLib.Name
	if l.Libraries[name] == nil {
		l.Libraries[name] = indexLib.ExtractLibrary()
	} else {
		release := indexLib.ExtractRelease()
		lib := l.Libraries[name]
		lib.Releases = append(lib.Releases, release)
		if lib.Latest.Version < release.Version {
			lib.Latest = release
		}
	}
}

// Names returns an array with all the names of the registered libraries
func (l *StatusContext) Names() []string {
	res := make([]string, len(l.Libraries))
	i := 0
	for n := range l.Libraries {
		res[i] = n
		i++
	}
	sortutil.CiAsc(res)
	return res
}

func CreateStatusContextFromIndex(index *Index, sketchbookPaths []string, corePaths []string) (*StatusContext, error) {
	// Start with an empty status context
	libraries := StatusContext{
		Libraries: map[string]*Library{},
	}
	for _, lib := range index.Libraries {
		// Add all indexed libraries in the status context
		libraries.AddLibrary(&lib)
	}
	return &libraries, nil
}
