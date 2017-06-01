package libraries

import (
	"encoding/json"
	"io/ioutil"
)

// Index represents the content of a library_index.json file
type Index struct {
	Libraries []IndexRelease
}

// IndexRelease is an entry of a library_index.json
type IndexRelease struct {
	Name            string
	Version         string
	Author          string
	Maintainer      string
	Sentence        string
	Paragraph       string
	Website         string
	Category        string
	Architectures   []string
	Types           []string
	URL             string
	ArchiveFileName string
	Size            int
	Checksum        string
}

// LoadLibrariesIndex reads a library_index.json from a file and returns
// the corresponding LibrariesIndex structure
func LoadLibrariesIndex(libFile string) (*Index, error) {
	libBuff, err := ioutil.ReadFile(libFile)
	if err != nil {
		return nil, err
	}

	var index Index
	if err := json.Unmarshal(libBuff, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// ExtractRelease create a new Release with the information contained
// in this index element
func (indexLib *IndexRelease) ExtractRelease() *Release {
	return &Release{
		Version:         indexLib.Version,
		URL:             indexLib.URL,
		ArchiveFileName: indexLib.ArchiveFileName,
		Size:            indexLib.Size,
		Checksum:        indexLib.Checksum,
	}
}

// ExtractLibrary create a new Library with the information contained
// in this index element
func (indexLib *IndexRelease) ExtractLibrary() *Library {
	release := indexLib.ExtractRelease()
	return &Library{
		Name:          indexLib.Name,
		Author:        indexLib.Author,
		Maintainer:    indexLib.Maintainer,
		Sentence:      indexLib.Sentence,
		Paragraph:     indexLib.Paragraph,
		Website:       indexLib.Website,
		Category:      indexLib.Category,
		Architectures: indexLib.Architectures,
		Types:         indexLib.Types,
		Releases:      []*Release{release},
		Latest:        release,
	}
}
