package goini

import (
	"path/filepath"
	"errors"
	"log"
)

// Suppress error if they are not otherwise used.
var _ = log.Printf

const (
	InheritedFrom = "inherited_from" // The key of the INI path which will be inherited from
)

// LoadInheritedINI loads an INI file which inherits from another INI
// e.g:
//	The common.ini has contents:
//		project=common
//		ip=192.168.0.1
//
//	And the project.ini has contents:
//		project=ppp
//		combo=ppp
//		inherited_from=common.ini
//
//	The project.ini has the same configure as below :
//		project=ppp
//		combo=ppp
//		ip=192.168.0.1
//
func LoadInheritedINI(filename string) (*INI, error) {
	ini := New()
	err := ini.ParseFile(filename)
	if err != nil {
		return nil, err
	}
	
	inherited, ok := ini.Get(InheritedFrom)
	if !ok {
		return ini, nil
	}
	
	inherited = GetPathByRelativePath(filename, inherited)
	inheritedINI, err := LoadInheritedINI(inherited)
	if err != nil {
		return nil, errors.New(err.Error() + " " + inherited)
	}
	
	ini.Merge(inheritedINI, false)
	return ini, nil
}

// Merge merges the data in another INI (from) to this INI (ini), and
// from INI will not be changed
func (ini *INI) Merge(from *INI, override bool) {
	for section, kv := range from.sections {
		for key, value := range kv {
			_, found := ini.SectionGet(section, key)
			if override || !found {
				ini.SectionSet(section, key, value)
			}
		}
	}
}

// GetPathByRelativePath gets the real path according to the relative file path
// e.g. :
// 	relativeFilePath = /home/goini/conf/common.conf
//	inheritedPath = app.conf
//
// and then the GetPathByRelativePath(relativeFilePath, inheritedPath) will
// return /home/goini/conf/app.conf
func GetPathByRelativePath(relativeFilePath, inheritedPath string) string {
	if filepath.IsAbs(inheritedPath) {
		return inheritedPath
	}
	
	dir, _ := filepath.Split(relativeFilePath)
	return filepath.Join(dir, inheritedPath)
}